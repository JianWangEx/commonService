// Package generate_model @Author  wangjian    2023/6/15 12:02 PM
package generate_model

import (
	"bytes"
	"fmt"
	"github.com/JianWangEx/commonService/util"
	"go/format"
	"strings"
	"text/template"
)

type ModelOutputFormatter struct {
	ModuleName         string
	PackagePath        string   // 包路径
	NeedImportPkgPaths []string // 需要导入的包路径
	TableNames         []string
	TablesColumns      map[string][]TableColumn
	OutputBytes        map[string][]byte
}

func GetTableStructName(tableName string) string {
	return util.ConvertToCamelCase(tableName, true)
}

func GetColumnDefinition(tc TableColumn) string {
	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("%s %s `gorm:\"column:%s", util.ConvertToCamelCase(tc.ColumnName, true), util.ConvertDBTypeToGolang(tc.DataType), tc.ColumnName))
	if tc.ColumnKey == "PRI" {
		builder.WriteString(fmt.Sprintf(";primaryKey"))
	}
	if tc.ColumnDefault != "" {
		builder.WriteString(fmt.Sprintf(";default:%s", tc.ColumnDefault))
	}
	if tc.IsNullable == "NO" {
		builder.WriteString(fmt.Sprintf(";not null"))
	}
	builder.WriteString(fmt.Sprintf("\" json:\"%s\"`", tc.ColumnName))
	return builder.String()
}

func NewFormatFileHelpers(mf *ModelOutputFormatter) []FormatFileHelper {
	var helpers = make([]FormatFileHelper, 0)
	for _, tableName := range mf.TableNames {
		var builder strings.Builder
		for _, column := range mf.TablesColumns[tableName] {
			columnDefinition := GetColumnDefinition(column)
			builder.WriteString(columnDefinition)
			builder.WriteString("\n")
		}

		helper := FormatFileHelper{
			PkgName:                util.GetPackageNameFromPackageFullPath(mf.PackagePath),
			NeedImportPkgPaths:     NormalizeNeedImportPaths(mf.NeedImportPkgPaths),
			TableStructName:        GetTableStructName(tableName),
			TableName:              tableName,
			TableColumnsDefinition: builder.String(),
		}

		helpers = append(helpers, helper)
	}

	return helpers
}

type FormatFileHelper struct {
	PkgName                string
	NeedImportPkgPaths     string
	TableStructName        string
	TableName              string
	TableColumnsDefinition string
}

func (f *ModelOutputFormatter) Format() error {
	tmpl := template.Must(template.New("modelFileTemplate").Parse(modelFileTemplate))
	formatFileHelpers := NewFormatFileHelpers(f)
	for _, helper := range formatFileHelpers {
		var buf bytes.Buffer
		err := tmpl.Execute(&buf, helper)
		if err != nil {
			return err
		}
		formatted, err := format.Source(buf.Bytes())
		if err != nil {
			return err
		}
		f.OutputBytes[helper.TableName] = formatted
	}
	return nil
}

func NormalizeNeedImportPaths(paths []string) string {
	var builder strings.Builder
	for _, path := range paths {
		builder.WriteString(fmt.Sprintf("\"%s\"\n", path))
	}
	return builder.String()
}
