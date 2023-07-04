// Package util @Author  wangjian    2023/7/3 1:18 PM
package util

import (
	"fmt"
	"github.com/JianWangEx/commonService/constant"
	"testing"
	"time"
)

type testSheet struct {
	Name      string
	Age       int
	IsActive  bool
	StartTime int64
}

func (s *testSheet) GetFieldVerificationMapping() map[string][]string {
	return map[string][]string{
		"StartTime": []string{"year", "month", "day", "hour", "minute", "second"},
	}
}

func (s *testSheet) GetCombinedFieldsCalculateFunction() CalculateFunction {
	return func(params ...interface{}) error {
		if len(params) != 6 {
			return constant.ErrorParamNumberIncorrect
		}
		year, ok := params[0].(int)
		if !ok {
			return constant.ErrorInvalidParamType
		}
		month, ok := params[1].(time.Month)
		if !ok {
			return constant.ErrorInvalidParamType
		}
		day, ok := params[2].(int)
		if !ok {
			return constant.ErrorInvalidParamType
		}
		hour, ok := params[3].(int)
		if !ok {
			return constant.ErrorInvalidParamType
		}
		minute, ok := params[4].(int)
		if !ok {
			return constant.ErrorInvalidParamType
		}
		second, ok := params[5].(int)
		if !ok {
			return constant.ErrorInvalidParamType
		}
		s.StartTime = time.Date(year, month, day, hour, minute, second, 0, time.Local).UnixMilli()
		return nil
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
