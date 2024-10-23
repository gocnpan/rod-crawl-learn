package main

import (
	"errors"
	"math/rand"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
	"github.com/go-rod/rod/lib/utils"
	"github.com/otiai10/copy"
)

var (
	baseUrl        = "https://learn.lianglianglee.com"
	baseOfflineDir string
	staticDir      = "./static"
	rng            *rand.Rand
)

func init() {
	baseOfflineDir = filepath.Join(baseDir, "offline")
	err := MakeAllDirIfNotExist(baseOfflineDir)
	if err != nil {
		panic("创建offline文件夹失败")
	}

	// 使用当前时间的纳秒级时间戳作为种子来创建一个新的随机数生成器
	source := rand.NewSource(time.Now().UnixNano())
	rng = rand.New(source)
}

// 爬取专栏页面
// 获取到各技术摘抄目录页面的url
func crawlColumns() {
	log.Debug("开始爬取专栏页面")
	page := browser.MustPage("https://learn.lianglianglee.com/%e4%b8%93%e6%a0%8f")
	defer page.MustClose()

	lis := page.MustWaitStable().
		MustElement("body > div > div.off-canvas-content > div.columns > div > div.book-content > div.book-post > div:nth-child(4) > ul").
		MustElements("li")

	rows := make([]TableColumn, 0, len(lis))
	for _, li := range lis {
		a := li.MustElement("a")
		txt := a.MustText()
		url := a.MustAttribute("href")
		rows = append(rows, TableColumn{
			Href:   *url,
			Name:   txt,
			Status: CrawlStatusInit,
		})
	}

	err := InsertColumn(rows)
	if err != nil {
		log.Panicf("新增专栏记录失败: %s", err)
	}

	log.Debug("爬取专栏页面完成")
}

// 爬取专栏页面的课程url
func crawlColCourUrls() {
	cols, err := SelectUnCrawledColumn()
	if err != nil {
		log.Panicf("查询未爬取的专栏记录失败: %s", err)
	}
	log.Debugf("开始爬取专栏页面的课程url, 长度: %d", len(cols))
	for _, col := range cols {
		err = makeCourseDir(col)
		if err != nil {
			continue
		}
		crawlCourseHref(col)
		sleepMin3()
	}
	log.Debug("爬取专栏页面的课程url完成")
}

// 爬取课程的url
func crawlCourseHref(column TableColumn) {
	u := baseUrl + column.Href
	log.Debugf("开始爬取课程(%s)href, url: %s", column.Name, u)
	// page := browser.MustPage(u)
	page, err := browser.Page(proto.TargetCreateTarget{URL: u})
	if err != nil {
		log.Warnf("爬取课程(%s)url列表失败失败: %s", column.Href, err)
		UpdateCourseStatus(column.Href, CrawlStatusErr)
		return
	}
	defer page.MustClose()

	ul, err := page.MustWaitStable().
		Element("body > div > div.book-sidebar > div.book-menu.uncollapsible > ul:nth-child(2)")
		// Element("body > div > div.off-canvas-content > div.columns > div > div.book-content > div.book-post > div:nth-child(4) > ul")
	if err != nil {
		log.Warnf("获取课程(%s)href失败: %s", column.Name, err)
	}
	if ul == nil {
		log.Warnf("获取课程(%s)ul失败", column.Name)
	}
	if err != nil || ul == nil {
		UpdateColumnStatus(column.Href, CrawlStatusErr)
		return
	}

	lis := ul.MustElements("li")

	rows := make([]TableCourse, 0, len(lis))
	for _, li := range lis {
		a := li.MustElement("a")
		txt := a.MustText()
		url := a.MustAttribute("href")
		rows = append(rows, TableCourse{
			Href:    *url,
			Column:  column.Name,
			Chapter: strings.TrimSuffix(txt, ".md"),
			Status:  CrawlStatusInit,
		})
	}
	err = InsertCourse(rows)
	if err != nil {
		log.Warnf("新增课程记录失败: %s", err)
		UpdateColumnStatus(column.Href, CrawlStatusErr)
	} else {
		log.Debugf("爬取课程(%s)href完成, len: %d", column.Name, len(rows))
		UpdateColumnStatus(column.Href, CrawlStatusOK)
	}
}

// 准备基础文件(夹)
func makeCourseDir(col TableColumn) error {
	dir := filepath.Join(baseOfflineDir, col.Name)
	err := MakeAllDirIfNotExist(filepath.Join(dir, "assets"))
	if err != nil {
		log.Warnf("创建课程(%s)文件夹失败: %s", col.Name, err)
		return err
	}
	err = copy.Copy(staticDir, filepath.Join(dir, "static"))
	if err != nil {
		log.Warnf("复制静态文件到课程(%s)文件夹失败: %s", col.Name, err)
		return err
	}
	return nil
}

// 爬取页面的内容
func crawlCoursePages() {
	courses, err := SelectUnCrawledCourse()
	if err != nil {
		log.Panicf("查询未爬取的课程记录失败: %s", err)
	}
	log.Debugf("开始爬取页面的内容, 长度: %d", len(courses))
	for _, course := range courses {
		// crawlCoursePage(course)
		timeCtrl(course, 0)
		sleepMin3()
	}
	log.Debug("爬取页面的内容完成")
}

// 这里增加了超时控制
// 当获取一个课程内容超过5分钟，则激活超时动作，重新获取该课程
// 如果超过3次均超时则放弃该课程
func timeCtrl(course TableCourse, times int) {
	ch := make(chan struct{}, 1)
	och := make(chan struct{}, 1)

	go func(och chan struct{}) {
		time.Sleep(5 * time.Minute)
		och <- struct{}{}
		close(och)
	}(och)
	go crawlCoursePage(course, ch)

	select {
	case <-ch:
		return
	case <-och:
		log.Warnf("页面(%s)爬取超时 %d 次", course.Href, times+1)
		if times > 2 {
			return
		}
		times++
		timeCtrl(course, times)
		return
	}
}

func crawlCoursePage(course TableCourse, ch chan struct{}) {
	defer func() {
		ch <- struct{}{}
		close(ch)
	}()

	log.Debugf("开始爬取页面(%s)的内容", course.Href)
	// page:= browser.MustPage(baseUrl + course.Href)
	page, err := browser.Page(proto.TargetCreateTarget{URL: baseUrl + course.Href})
	if err != nil {
		log.Warnf("打开页面(%s)失败: %s", course.Href, err)
		UpdateCourseStatus(course.Href, CrawlStatusErr)
		return
	}
	defer page.MustClose()

	// 如果 不是 .md 页面则直接写 html
	if !strings.HasSuffix(course.Href, ".md") {
		log.Warnf("页面(%s)不是 .md 页面，直接写 html", course.Href)
		err := saveHtml(page, course)
		if err != nil {
			log.Warnf("保存页面(%s)内容失败: %s", course.Href, err)
			UpdateCourseStatus(course.Href, CrawlStatusErr)
		} else {
			log.Debugf("爬取页面(%s)的内容完成", course.Href)
			UpdateCourseStatus(course.Href, CrawlStatusOK)
		}
		return
	}

	err = saveImgs(page, course)
	if err != nil {
		UpdateCourseStatus(course.Href, CrawlStatusErr)
		return
	}

	err = saveMd(page, course)
	if err != nil {
		UpdateCourseStatus(course.Href, CrawlStatusErr)
		return
	}

	UpdateCourseStatus(course.Href, CrawlStatusOK)
	log.Debugf("爬取页面(%s)的内容完成", course.Href)
}

func saveImgs(page *rod.Page, course TableCourse) error {
	err := MakeAllDirIfNotExist(filepath.Join(baseOfflineDir, course.Column, "assets"))
	if err != nil {
		log.Panicf("创建assets文件夹失败: %v", err)
	}
	contents, err := page.
		MustWaitLoad().
		Element("body > div > div.off-canvas-content > div.columns > div > div.book-content > div.book-post > div:nth-child(4)")
	// 这里表明，网页已经无法访问到了：被反爬限制
	if err != nil {
		log.Warnf("页面(%s)可能被反爬: %s", course.Href, err)
		return err
	}
	// 这里表明，网页已经无法访问到了：被反爬限制
	if contents == nil {
		log.Warnf("页面(%s)可能被反爬", course.Href)
		return errors.New("页面可能被反爬")
	}

	imgs, err := contents.Elements("img")
	if err != nil {
		log.Warnf("获取课程(%s)图片失败: %s", course.Href, err)
		return err
	}
	// 创建文件夹

	for _, img := range imgs {
		ip := filepath.Join(baseOfflineDir, course.Column, *img.MustAttribute("src"))
		bt := img.MustResource()
		err = utils.OutputFile(ip, bt)
		if err != nil {
			log.Warnf("保存图片(image path: %s)失败: %v", ip, err)
			return err
		}
	}
	log.Debugf("爬取课程(%s)图片(len: %d)完成", course.Chapter, len(imgs))
	return nil
}

func saveMd(page *rod.Page, course TableCourse) error {
	modList(page)
	addCss(page)
	return saveHtml(page, course)
}

// 修改目录索引值
func modList(page *rod.Page) {
	// 删除
	page.MustElement("body > div > div.book-sidebar > div.book-brand").Remove()

	// 替换目录
	ats := page.MustElements("body > div > div.book-sidebar > div.book-menu.uncollapsible > ul:nth-child(2) > li > a")
	// log.Debugf("开始修改目录, a tags len: %d", len(ats))
	ule := `<ul class="uncollapsible">
	`
	for _, a := range ats {
		at := a.MustText()
		// log.Debugf("读取到目录: %s", at)
		nat := strings.TrimSuffix(at, ".md")
		href := "./" + nat + ".html"
		ule += `<li><a class="menu-item" href="` + href + `">` + nat + `</a></li>
		`
	}
	ule += "</ul>"
	page.MustEval(`() => {
		document.querySelector('body > div > div.book-sidebar > div.book-menu.uncollapsible').innerHTML = arguments[0];
	}`, ule)
}

func addCss(page *rod.Page) {
	// 添加样式文件
	css := `<link rel="stylesheet" href="./static/header.css">
	<link rel="stylesheet" href="./static/highlight.min.css">
	<link rel="stylesheet" href="./static/index.css">`
	page.MustEval(`() => {
		document.head.innerHTML += arguments[0];
	}`, css)
}

func saveHtml(page *rod.Page, course TableCourse) error {
	log.Debug("开始保存html")
	html := page.MustHTML()
	err := utils.OutputFile(
		filepath.Join(
			baseOfflineDir,
			course.Column,
			strings.TrimSuffix(course.Chapter, ".md")+".html"),
		[]byte(html),
	)
	if err != nil {
		log.Warnf("保存课程(%s)html失败: %v", course.Chapter, err)
		return err
	}
	log.Debugf("保存课程(%s)html成功", course.Chapter)
	return nil
}

func sleepMin3() {
	sleep(3)
}

func sleepMin5() {
	sleep(5)
}

func sleep(base int) {
	sleepTime := rng.Intn(8) + base
	log.Debugf("暂停 %d 秒", sleepTime)
	time.Sleep(time.Duration(sleepTime) * time.Second)
}
