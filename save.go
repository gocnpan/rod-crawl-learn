package main

import (
	"path/filepath"
	"strings"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/utils"
)


func init() {
}

// https://learn.lianglianglee.com/%E4%B8%93%E6%A0%8F/DDD%E5%AE%9E%E6%88%98%E8%AF%BE/00%20%E5%BC%80%E7%AF%87%E8%AF%8D%20%20%E5%AD%A6%E5%A5%BD%E4%BA%86DDD%EF%BC%8C%E4%BD%A0%E8%83%BD%E5%81%9A%E4%BB%80%E4%B9%88%EF%BC%9F.md
// https://learn.lianglianglee.com/%E4%B8%93%E6%A0%8F/DDD%E5%AE%9E%E6%88%98%E8%AF%BE/assets/dc32e8e4a317fe00121ce18adc407c66.jpg
func savePage() {
	page := browser.MustPage("https://learn.lianglianglee.com/%E4%B8%93%E6%A0%8F/DDD%20%E5%BE%AE%E6%9C%8D%E5%8A%A1%E8%90%BD%E5%9C%B0%E5%AE%9E%E6%88%98/00%20%E5%BC%80%E7%AF%87%E8%AF%8D%20%20%E8%AE%A9%E6%88%91%E4%BB%AC%E6%8A%8A%20DDD%20%E7%9A%84%E6%80%9D%E6%83%B3%E7%9C%9F%E6%AD%A3%E8%90%BD%E5%9C%B0.md")

	saveImgs(page)
	saveHtml(page)
}

func saveImgs(page *rod.Page) {
	err := MakeAllDirIfNotExist(filepath.Join(baseDir, "assets"))
	if err != nil {
		log.Panicf("创建assets文件夹失败: %v", err)
	}
	contents, err := page.MustWaitLoad().Element("body > div > div.off-canvas-content > div.columns > div > div.book-content > div.book-post > div:nth-child(4)")
	// 这里表明，网页已经无法访问到了：被反爬限制
	if err != nil {

	}
	// 这里表明，网页已经无法访问到了：被反爬限制
	if contents == nil {
		log.Warnf("没有找到内容")
	}
	imgs := page.MustWaitLoad().
		MustElement("body > div > div.off-canvas-content > div.columns > div > div.book-content > div.book-post > div:nth-child(4)").
		MustElements("img")

		// 创建文件夹

	for _, img := range imgs {
		ip := filepath.Join(baseDir, *img.MustAttribute("src"))
		bt := img.MustResource()
		err = utils.OutputFile(ip, bt)
		if err != nil {
			log.Debugf("保存图片(image path: %s)失败: %v", ip, err)
		}
	}
}

func saveHtml(page *rod.Page) {
	log.Debug("开始修改目录链接")
	modList(page)
	log.Debug("开始增加css文件")
	addCss(page)

	log.Debug("开始保存html")
	html := page.MustHTML()
	err := utils.OutputFile("./data/test.html", []byte(html))
	if err != nil {
		log.Panicf("保存html失败: %v", err)
	}
	log.Debug("保存html成功")
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
		log.Debugf("读取到目录: %s", at)
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
