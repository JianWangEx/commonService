// Package util @Author  wangjian    2023/6/15 2:46 PM
package util

import "strings"

// ConvertDBTypeToGolang
//
//	@Description: 转换DB数据类型为Go支持的类型
//	@param t DB的类型，MySQL类型数据库
//	@return string Go支持的类型，如果出错，返回为空
func ConvertDBTypeToGolang(t string) string {
	t = strings.ToLower(t)

	switch t {
	case "varchar", "char", "text":
		return "string"
	case "int", "integer", "tinyint", "smallint", "mediumint":
		return "int"
	case "bigint":
		return "int64"
	case "float", "double", "decimal":
		return "float64"
	case "boolean":
		return "bool"
	case "date", "datetime", "timestamp":
		return "time.Time"
	default:
		return ""
	}
}
