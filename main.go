package main

import (
	"runtime/debug"

	"github.com/go-rod/rod"
)

var (
	browser *rod.Browser
	baseDir = "./data"
	log     = GetLogger()
)

func init() {
	err := MakeAllDirIfNotExist(baseDir)
	if err != nil {
		panic("创建data文件夹失败")
	}
	RunSQLite()
}

func main() {
	// 初始化浏览器连接
	browser = rod.New().MustConnect()
	defer func() {
		stack := string(debug.Stack())
		if r := recover(); r != nil {
			log.Warnf("panic: %v, stack: %s", r, stack)
		}
		browser.MustClose()
	}()

	// 爬取专栏href
	crawlColumns()
	sleepMin5()
	// 爬取课程href
	crawlColCourUrls()
	// 爬取课程内容
	crawlCoursePages()
}
