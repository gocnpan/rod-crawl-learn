package main

type Column struct {
	Href string `json:"href" gorm:"column:href;primaryKey"`
	Name string `json:"name" gorm:"column:name"`
	// 这里的抓取状态仅仅指对课程列表的抓取，并不抓取课程详情页
	Status int `json:"status" gorm:"column:status"` // 抓取状态 0:未爬取, 100:正在爬取, 200:已爬取完, 400:爬取出错
}

type Course struct {
	Href    string `json:"href" gorm:"column:href;primaryKey"` // 章节 href
	Column  string `json:"column" gorm:"column:column"`        // 课程名
	Chapter string `json:"chapter" gorm:"column:chapter"`      // 章节名
	Status  int    `json:"status" gorm:"column:status"`        // 抓取状态 0:未爬取, 100:正在爬取, 200:已爬取完, 400:爬取出错
}
