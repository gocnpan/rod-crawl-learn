package main

import "gorm.io/gorm/clause"

// 新增专栏
func InsertColumn(rows []TableColumn) error {
	return db.
		Clauses(clause.OnConflict{DoNothing: true}).
		Create(&rows).Error
}

// 修改专栏状态
func UpdateColumnStatus(href string, status int) error {
	return db.
		Model(&TableColumn{}).
		Where("href = ?", href).
		Update("status", status).
		Error
}

// 查询未爬取成功的专栏
func SelectUnCrawledColumn() ([]TableColumn, error) {
	var rows []TableColumn
	err := db.
		Where("status != ?", CrawlStatusOK).
		Order("created_at").
		Find(&rows).Error
	return rows, err
}

// 新增课程表
func InsertCourse(rows []TableCourse) error {
	return db.
		Clauses(clause.OnConflict{DoNothing: true}).
		Create(&rows).Error
}

// 修改课程状态
func UpdateCourseStatus(href string, status int) error {
	return db.
		Model(&TableCourse{}).
		Where("href = ?", href).
		Update("status", status).
		Error
}

// 查询未爬取成功的课程
func SelectUnCrawledCourse() ([]TableCourse, error) {
	var rows []TableCourse
	err := db.
		Where("status != ?", CrawlStatusOK).
		Order("created_at").
		Find(&rows).Error
	return rows, err
}
