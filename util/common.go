// Package util @Author  wangjian    2023/6/15 10:17 AM
package util

import (
	"strings"
)

// ConvertToCamelCase
//
//	@Description: 将下划线命名转为驼峰式
//	@param input
//	@param capitalizeFirst 首字母是否大写
//	@return string
func ConvertToCamelCase(input string, capitalizeFirst bool) string {
	// 将下划线替换为空格
	words := strings.Split(input, "_")

	// 将单词首字母大写并拼接
	var result string
	for i, word := range words {
		if len(word) == 0 {
			continue
		}
		if i == 0 && !capitalizeFirst {
			result += word
		} else {
			result += strings.Title(word)
		}
	}

	if len(result) == 0 {
		return ""
	}

	return result
}

// GetPackageNameFromPackageFullPath
//
//	@Description: 从包完整路径中获取包名
//	@param path
//	@return string 如果path不包含"/",则返回path
func GetPackageNameFromPackageFullPath(path string) string {
	idx := strings.LastIndex(path, "/")
	if idx == -1 {
		return path
	}
	return path[idx+1:]
}
