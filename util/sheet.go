// Package util @Author  wangjian    2023/7/2 11:56 PM
package util

import (
	"github.com/JianWangEx/commonService/constant"
	"github.com/xuri/excelize/v2"
	"reflect"
	"strconv"
	"strings"
)

type XLSLSheetConfig struct {
	FilePath  string
	SheetName string
}

type CalculateFunction func(paramsMapping map[string]interface{}) error

type SheetHandler interface {
	GetFieldVerificationMapping() map[string][]string
	GetCombinedFieldsCalculateFunction() map[string]CalculateFunction
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

	// 校验receiver是否为指向结构体数组的指针
	err = checkType(receiver)
	if err != nil {
		return err
	}

	rv := reflect.ValueOf(receiver)
	receiverSliceValue := rv.Elem()
	entityValue := reflect.New(receiverSliceValue.Type().Elem())

	// 检查属性行
	rowValueIndexMapping, fieldColumnsMapping, err := preCheck(rows[0], entityValue)
	if err != nil {
		return err
	}

	// 将每一行数据写入receiverSliceValue
	for _, row := range rows {
		newValue, err := writeData(rowValueIndexMapping, row, entityValue.Type(), fieldColumnsMapping)
		if err != nil {
			return err
		}
		receiverSliceValue = reflect.Append(receiverSliceValue, newValue)
	}
	rv.Set(receiverSliceValue)

	reflect.ValueOf(receiver).Elem().Set(receiverSliceValue)
	return nil
}

func preCheck(columns []string, value reflect.Value) (map[string]int, map[string][]string, error) {
	// step1. 生成sheet表列名->下标值映射
	rowValueIndexMapping := make(map[string]int, len(columns))
	for i, row := range columns {
		if _, ok := rowValueIndexMapping[row]; !ok {
			return nil, nil, constant.ErrorSheetAttributeRepeat
		}
		rowValueIndexMapping[row] = i
	}

	fieldColumnsMapping := make(map[string][]string)
	// step2. 检查value是否实现SheetHandler接口
	iType := reflect.TypeOf((*SheetHandler)(nil)).Elem()
	if value.Type().Implements(iType) {
		method := value.MethodByName("GetFieldVerificationMapping")
		f := method.Call(nil)
		fieldColumnsMapping = f[0].Interface().(map[string][]string)
	}

	// step3. 检查value字段值是否匹配
	vt := value.Type()
	for i := vt.NumField() - 1; i >= 0; i-- {
		fieldName := vt.Field(i).Name
		if _, ok := rowValueIndexMapping[fieldName]; !ok {
			if cs, ok := fieldColumnsMapping[fieldName]; ok {
				for _, c := range cs {
					if _, ok := rowValueIndexMapping[c]; !ok {
						return nil, nil, constant.ErrorDataStructNotMatch
					}
				}
			} else {
				return nil, nil, constant.ErrorDataStructNotMatch
			}
		}
	}

	return rowValueIndexMapping, fieldColumnsMapping, nil
}

func checkType(receiver interface{}) error {
	if receiver == nil {
		return constant.ErrorTypeNotPtrOrIsNil
	}

	rv := reflect.TypeOf(receiver)
	if rv.Kind() != reflect.Ptr {
		return constant.ErrorTypeNotPtrOrIsNil
	}

	sliceType := rv.Elem()
	if sliceType.Kind() != reflect.Slice {
		return constant.ErrorTypeIsNotSlice
	}

	structType := sliceType.Elem()
	if structType.Kind() != reflect.Struct {
		return constant.ErrorSliceDataTypeIsNotStruct
	}

	return nil
}

func writeData(rowValueIndexMapping map[string]int, data []string, t reflect.Type, fieldColumnsMapping map[string][]string) (reflect.Value, error) {
	res := reflect.New(t).Elem()
	fieldFunctionMapping := make(map[string]CalculateFunction)
	if len(fieldColumnsMapping) > 0 {
		method := res.MethodByName("GetCombinedFieldsCalculateFunction")
		f := method.Call(nil)
		fieldFunctionMapping = f[0].Interface().(map[string]CalculateFunction)
	}
	for i := res.NumField() - 1; i >= 0; i-- {
		rowIndex, ok := rowValueIndexMapping[t.Field(i).Name]
		if !ok {
			columns := fieldColumnsMapping[t.Field(i).Name]
			columnValueMapping := make(map[string]interface{})
			for _, column := range columns {
				fName, fType := getColumnValueAndType(column)
				if ri, ok := rowValueIndexMapping[fName]; !ok {
					return reflect.Value{}, constant.ErrorDataStructNotMatch
				} else {
					rowValue, err := AssertValue(data[ri], fType)
					if err != nil {
						return reflect.Value{}, err
					}
					columnValueMapping[fName] = rowValue
				}
			}
			fn := fieldFunctionMapping[t.Field(i).Name]
			err := fn(columnValueMapping)
			if err != nil {
				return reflect.Value{}, err
			}
			continue
		}
		v := data[rowIndex]
		fieldValue := reflect.New(res.Field(i).Type())
		switch fieldValue.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			tem, err := strconv.ParseInt(v, 10, 64)
			if err != nil {
				return reflect.Value{}, err
			}
			fieldValue.SetInt(tem)
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			tem, err := strconv.ParseUint(v, 10, 64)
			if err != nil {
				return reflect.Value{}, err
			}
			fieldValue.SetUint(tem)
		case reflect.String:
			fieldValue.SetString(v)
		case reflect.Bool:
			tmp, err := strconv.ParseBool(v)
			if err != nil {
				return reflect.Value{}, err
			}
			fieldValue.SetBool(tmp)
		case reflect.Float32, reflect.Float64:
			tmp, err := strconv.ParseFloat(v, 64)
			if err != nil {
				return reflect.Value{}, err
			}
			fieldValue.SetFloat(tmp)
		default:
			return reflect.Value{}, constant.ErrorStructDataTypeNotSupported
		}
		res.Field(i).Set(fieldValue)
	}
	return res, nil
}

func getColumnValueAndType(s string) (string, string) {
	sArr := strings.Split(s, "_")
	return sArr[0], sArr[1]
}
