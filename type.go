package main

const (
	CrawlStatusInit = 0 // 未爬取
	CrawlStatusOK   = 200
	CrawlStatusErr  = 400
)

type Time struct {
	CreatedAt int64 `json:"created_at,omitempty" gorm:"autoCreateTime:milli"`
	UpdatedAt int64 `json:"updated_at,omitempty" gorm:"autoUpdateTime:milli"`
}

type TableColumn struct {
	Href string `json:"href" gorm:"column:href;primaryKey"`
	Name string `json:"name" gorm:"column:name"`
	// 这里的抓取状态仅仅指对课程列表的抓取，并不抓取课程详情页
	Status int `json:"status" gorm:"column:status"` // 抓取状态 0:未爬取, 200:已爬取完, 400:爬取出错

	Time
}

func (TableColumn) TableName() string {
	return "column"
}

type TableCourse struct {
	Href    string `json:"href" gorm:"column:href;primaryKey"` // 章节 href
	Column  string `json:"column" gorm:"column:column"`        // 课程名
	Chapter string `json:"chapter" gorm:"column:chapter"`      // 章节名
	Status  int    `json:"status" gorm:"column:status"`        // 抓取状态 0:未爬取, 200:已爬取完, 400:爬取出错

	Time
}

func (TableCourse) TableName() string {
	return "course"
}
