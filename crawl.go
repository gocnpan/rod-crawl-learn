package main


// 爬取专栏页面
// 获取到各技术摘抄目录页面的url
func crawlHome() {
	page := browser.MustPage("https://learn.lianglianglee.com/%e4%b8%93%e6%a0%8f")
	lis := page.MustWaitStable().
		MustElement("body > div > div.off-canvas-content > div.columns > div > div.book-content > div.book-post > div:nth-child(4) > ul").
		MustElements("li")

	for _, li := range lis {
		a := li.MustElement("a")
		txt := a.MustText()
		url := a.MustAttribute("href")
		log.Debugf("%s: %s", txt, *url)
	}
	page.Activate()
}
