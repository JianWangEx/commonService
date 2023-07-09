// Package util @Author  wangjian    2023/7/3 1:18 PM
package util

import (
	"fmt"
	"github.com/JianWangEx/commonService/constant"
	"reflect"
	"testing"
	"time"
)

type testSheet struct {
	Name      string
	Age       int
	IsActive  bool
	StartTime int64
}

func (s testSheet) GetFieldVerificationMapping() map[string][]string {
	return map[string][]string{
		"StartTime": []string{"year_int", "month_int", "day_int", "hour_int", "minute_int", "second_int"},
	}
}

func (s testSheet) GetCombinedFieldsCalculateFunction() map[string]CalculateFunction {
	return map[string]CalculateFunction{
		"StartTime": func(paramsMapping map[string]interface{}) (interface{}, error) {
			year, ok := paramsMapping["year"].(int)
			if !ok {
				return nil, constant.ErrorInvalidParamType
			}
			month, ok := paramsMapping["month"].(int)
			if !ok {
				return nil, constant.ErrorInvalidParamType
			}
			day, ok := paramsMapping["day"].(int)
			if !ok {
				return nil, constant.ErrorInvalidParamType
			}
			hour, ok := paramsMapping["hour"].(int)
			if !ok {
				return nil, constant.ErrorInvalidParamType
			}
			minute, ok := paramsMapping["minute"].(int)
			if !ok {
				return nil, constant.ErrorInvalidParamType
			}
			second, ok := paramsMapping["second"].(int)
			if !ok {
				return nil, constant.ErrorInvalidParamType
			}
			res := time.Date(year, time.Month(month), day, hour, minute, second, 0, time.Local).UnixMilli()
			return res, nil
		},
	}
}

func TestParseXLSLSheet(t *testing.T) {
	config := &XLSLSheetConfig{
		FilePath:  "/Users/wangjian/test_sheet.xlsx",
		SheetName: "Sheet1",
	}

	data := []testSheet{}

	err := ParseXLSLSheet(config, &data)
	if err != nil {
		panic(err)
	}
	fmt.Println(data)
}

func TestReflect(t *testing.T) {
	data := testSheet{}
	value := reflect.ValueOf(data)
	method := value.MethodByName("GetFieldVerificationMapping")
	res := method.Call(nil)
	fmt.Println(res[0].Interface().(map[string][]string))
}
