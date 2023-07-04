// Package util @Author  wangjian    2023/7/2 11:56 PM
package util

import (
	"github.com/JianWangEx/commonService/constant"
	"github.com/xuri/excelize/v2"
	"reflect"
	"strconv"
)

type XLSLSheetConfig struct {
	FilePath  string
	SheetName string
}

type CalculateFunction func(params ...interface{}) error

type SheetHandler interface {
	GetFieldVerificationMapping() map[string][]string
	GetCombinedFieldsCalculateFunction() CalculateFunction
}

// ParseXLSLSheet
//
//	@Description: 解析XLSL类型sheet文件数据到receiver
//	@param filePath sheet文件路径，比如/user/tab.xlsx
//	@param receiver 应该是结构体数组指针类型并且不为空
//	@return error
func ParseXLSLSheet(config *XLSLSheetConfig, receiver interface{}) error {
	// 打开XLSX文件，获取所有的工作表
	f, err := excelize.OpenFile(config.FilePath)
	if err != nil {
		panic(err)
	}

	// 获取工作表的行数据
	rows, err := f.GetRows(config.SheetName)
	if err != nil {
		panic(err)
	}

	// 检查属性行
	m, err := preCheck(rows[0], receiver)
	if err != nil {
		return err
	}

	rv := reflect.ValueOf(receiver)
	sv := rv.Elem()

	// 映射数据
	for i := 1; i < len(rows); i++ {
		row := rows[i]
		elementType := sv.Type().Elem()
		// 构造一个新的零值结构体对象
		obj := reflect.New(elementType).Elem()
		for i := elementType.NumField() - 1; i >= 0; i-- {
			field := elementType.Field(i)
			v := row[m[field.Name]]
			fieldValue := obj.FieldByName(field.Name)
			switch field.Type.Kind() {
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				tem, err := strconv.ParseInt(v, 10, 64)
				if err != nil {
					return err
				}
				fieldValue.SetInt(tem)
			case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				tem, err := strconv.ParseUint(v, 10, 64)
				if err != nil {
					return err
				}
				fieldValue.SetUint(tem)
			case reflect.String:
				fieldValue.SetString(v)
			case reflect.Bool:
				tmp, err := strconv.ParseBool(v)
				if err != nil {
					return err
				}
				fieldValue.SetBool(tmp)
			case reflect.Float32, reflect.Float64:
				tmp, err := strconv.ParseFloat(v, 64)
				if err != nil {
					return err
				}
				fieldValue.SetFloat(tmp)
			default:
				return constant.ErrorStructDataTypeNotSupported
			}
		}
		sv = reflect.Append(sv, obj)
	}

	reflect.ValueOf(receiver).Elem().Set(sv)
	return nil
}

func preCheck(raws []string, receiver interface{}) (map[string]int, error) {
	if receiver == nil {
		return nil, constant.ErrorTypeNotPtrOrIsNil
	}
	rv := reflect.TypeOf(receiver)
	if rv.Kind() != reflect.Ptr {
		return nil, constant.ErrorTypeNotPtrOrIsNil
	}
	sliceType := rv.Elem()
	// 检查receiver类型是否为slice
	if sliceType.Kind() != reflect.Slice {
		return nil, constant.ErrorTypeIsNotSlice
	}
	elementType := sliceType.Elem()
	// 检查receiver元素类型是否为结构体
	if elementType.Kind() != reflect.Struct {
		return nil, constant.ErrorSliceDataTypeIsNotStruct
	}

	if len(raws) != elementType.NumField() {
		return nil, constant.ErrorDataStructNotMatch
	}

	// 校验属性key是否相同
	m := make(map[string]int, len(raws))
	for i, raw := range raws {
		if _, ok := m[raw]; ok {
			return nil, constant.ErrorSheetAttributeRepeat
		}
		m[raw] = i
	}
	for i := elementType.NumField() - 1; i >= 0; i-- {
		if _, ok := m[elementType.Field(i).Name]; !ok {
			return nil, constant.ErrorDataStructNotMatch
		}
	}
	return m, nil
}
