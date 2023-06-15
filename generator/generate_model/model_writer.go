// Package generate_model @Author  wangjian    2023/6/15 3:57 PM
package generate_model

import (
	"fmt"
	"os"
	"strings"
)

type ModelWriter struct {
	packagePath string
	fileNames   []string
	tableNames  []string
	outputBytes map[string][]byte
}

func NewModelWriter(c *Config, outputBytes map[string][]byte) *ModelWriter {
	return &ModelWriter{
		packagePath: c.Path,
		fileNames:   c.FileNames,
		tableNames:  c.TableNames,
		outputBytes: outputBytes,
	}
}

func (w *ModelWriter) Write() error {
	// 如果package目录不存在创建目录
	err := w.MakeDir()
	if err != nil {
		panic(err)
	}
	for i, t := range w.tableNames {
		filePath := w.packagePath + "/" + w.fileNames[i]
		f, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
		if err != nil {
			return err
		}
		_, err = f.Write(w.outputBytes[t])
		if err != nil {
			return err
		}
		err = f.Close()
		if err != nil {
			fmt.Println(err.Error())
		}
	}

	return nil
}

func (w *ModelWriter) MakeDir() error {
	dirArr := strings.Split(w.packagePath, "/")
	var currDirStr string
	for _, dir := range dirArr {
		if dir == "" || dir == "." {
			continue
		}
		if currDirStr == "" {
			currDirStr = dir
		} else {
			currDirStr += "/" + dir
		}
		if _, err := os.Stat(currDirStr); os.IsNotExist(err) {
			fmt.Printf("\ncreating directory %v\n", currDirStr)
			err = os.Mkdir(currDirStr, 0755)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
