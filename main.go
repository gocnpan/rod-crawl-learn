package main

import (
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
}

func main() {
	// 初始化浏览器连接
	browser = rod.New().MustConnect()
	defer func() {
		if r := recover(); r != nil {
			log.Warnf("panic: %v", r)
		}
		browser.MustClose()
	}()
	// defer browser.MustClose()

	// crawlHome()
	savePage()
}
