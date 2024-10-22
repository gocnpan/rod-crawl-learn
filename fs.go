package main

import (
	"os"
)

// 只获取文件夹
// func GetOnlyDir(path string) string {
// 	paths := strings.Split(path, "/")
// 	if len(paths) > 1 {
// 		return strings.Join(paths[:len(paths)-2], "/")
// 	}
// 	return path
// }

// 创建文件夹
func MakeAllDirIfNotExist(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		err := os.MkdirAll(path, 0777)
		if err != nil {
			return err
		}
	}
	return nil
}
