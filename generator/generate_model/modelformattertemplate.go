// Package generate_model @Author  wangjian    2023/6/15 11:32 AM
package generate_model

// Templates

const modelFileTemplate = `
// CODE GENERATED AUTOMATICALLY BY github.com/JianWangEx/commonService/generator/generate_model
// THIS FILE SHOULD NOT BE EDITED MANUALLY

package {{.PkgName}}

import (
	{{.NeedImportPkgPaths}}
)

const TableName{{.TableStructName}} = "{{.TableName}}"

type {{.TableStructName}} struct {
	{{.TableColumnsDefinition}}
}

// TableName {{.TableStructName}}'s table name
func (t *{{.TableStructName}}) TableName() string {
	return TableName{{.TableStructName}}
}

func (t *{{.TableStructName}}) GetTableInfo() (string, string) {
	return config.GetDBName(), t.TableName()
}

`
