package stream

import "os"

// 如果文件夹不存在，则创建
func CreateMoreFolder(path string) (err error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		// 先创建文件夹  多级文件夹
		err := os.MkdirAll(path, os.ModePerm)
		// 单个文件夹
		// os.Mkdir(path, 0777)
		// 再修改权限
		// os.Chmod(path, 0777)

		return err
	}

	return err
}
