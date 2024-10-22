package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
	gormlogger "gorm.io/gorm/logger"
)

var Logger *zap.SugaredLogger

var loggerMutex sync.RWMutex

var levelMap = map[string]zapcore.Level{
	"debug":  zapcore.DebugLevel,
	"info":   zapcore.InfoLevel,
	"warn":   zapcore.WarnLevel,
	"error":  zapcore.ErrorLevel,
	"dpanic": zapcore.DPanicLevel,
	"panic":  zapcore.PanicLevel,
	"fatal":  zapcore.FatalLevel,
}

func GetLogger() *zap.SugaredLogger {
	loggerMutex.Lock()
	defer loggerMutex.Unlock()
	if Logger == nil {
		initLogger()
	}
	return Logger
}

func getLoggerLevel(lvl string) zapcore.Level {
	if level, ok := levelMap[lvl]; ok {
		return level
	}
	return zapcore.InfoLevel
}

// level: debug info warn error dpanic panic fatal
func initLogger() {
	hook := lumberjack.Logger{
		Filename:   filepath.Join(baseDir, "logs.log"),
		MaxSize:    10,
		MaxBackups: 10,
		MaxAge:     30,
		LocalTime:  true,
		Compress:   true,
	}
	var syncWrite zapcore.WriteSyncer
	// 在控制台输出
	syncWrite = zapcore.NewMultiWriteSyncer(zapcore.AddSync(os.Stdout), zapcore.AddSync(&hook))
	// 不在控制台输出
	// syncWrite = zapcore.AddSync(&hook)

	lvl := getLoggerLevel("debug")
	encoder := zap.NewProductionConfig()
	encoder.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	core := zapcore.NewCore(zapcore.NewJSONEncoder(encoder.EncoderConfig), syncWrite, zap.NewAtomicLevelAt(lvl))
	log := zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))
	Logger = log.Sugar()
}

// NewGormLogger 返回带 zap logger 的 GormLogger
func NewGormLogger(log *zap.SugaredLogger, slowThreshold time.Duration) GormLogger {
	st := slowThreshold
	if st == 0 {
		st = 200 * time.Millisecond
	}
	return GormLogger{
		log:           log,
		slowThreshold: st,
	}
}

type GormLogger struct {
	log           *zap.SugaredLogger
	slowThreshold time.Duration
}

// LogMode 实现 gorm logger 接口方法
func (g GormLogger) LogMode(gormLogLevel gormlogger.LogLevel) gormlogger.Interface {
	newlogger := g
	return newlogger
}

// Info 实现 gorm logger 接口方法
func (g GormLogger) Info(ctx context.Context, msg string, data ...interface{}) {
	g.log.Infof(msg, data...)
}

// Warn 实现 gorm logger 接口方法
func (g GormLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	g.log.Warnf(msg, data...)
}

// Error 实现 gorm logger 接口方法
func (g GormLogger) Error(ctx context.Context, msg string, data ...interface{}) {
	g.log.Errorf(msg, data...)
}

// Trace 实现 gorm logger 接口方法
func (g GormLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	elapsed := time.Since(begin)
	sql, rows := fc()
	sql = RemoveDuplicateWhitespace(sql, true)
	log := g.log

	rs := fmt.Sprintf("%v", rows)
	if rows == -1 {
		rs = fmt.Sprintf("%v", "-")
	}
	switch {
	case err != nil:
		log.Errorf("Error: %s, [%.3fms] [rows:%v]. SQL: %s.", err, float64(elapsed.Nanoseconds())/1e6, rs, sql)
	case g.slowThreshold != 0 && elapsed > g.slowThreshold:
		log.Warnf("SLOW SQL >= %v, [%.3fms] [rows:%v]. SQL: %s.", g.slowThreshold, float64(elapsed.Nanoseconds())/1e6, rs, sql)
	default:
		// log.Debugf("[%.3fms] [rows:%v]. SQL: %s.", float64(elapsed.Nanoseconds())/1e6, rs, sql)
	}
}

// RemoveDuplicateWhitespace 删除字符串中重复的空白字符为单个空白字符
// trim: 是否去掉首位空白
func RemoveDuplicateWhitespace(s string, trim bool) string {
	ws, err := regexp.Compile(`\s+`)
	if err != nil {
		return s
	}
	s = ws.ReplaceAllString(s, " ")
	if trim {
		s = strings.TrimSpace(s)
	}
	return s
}
