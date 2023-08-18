package generate_obj_and_mock_file

// Templates

const initFileTemplate = `
// CODE GENERATED AUTOMATICALLY BY github.com/JianWangEx/commonService/generator/generate_obj_and_mock_file
// THIS FILE SHOULD NOT BE EDITED MANUALLY

package {{.TypeStr}}

import (
	{{.PkgName}} "{{.PkgFullPath}}"
	"{$register.ModuleName}/pkg/ioc"
)

// nolint:funlen
func init() {

{{range $register := .Registers}}
	ioc.AddProvider({{$register.PackageName}}.{{$register.MethodName}})
{{end}}

}
`

const mockInitFileTemplate = `
// CODE GENERATED AUTOMATICALLY BY github.com/JianWangEx/commonService/generator/generate_obj_and_mock_file
// THIS FILE SHOULD NOT BE EDITED MANUALLY

package {{.TypeStr}}

import (
	{{.PkgName}} "{{.InterfaceFullPath}}"
	{{.PkgName}}mock "{{.PkgFullPath}}"
	"{{.ModuleName}}/pkg/ioc"
)

// nolint:funlen
func init() {
	{{range $register := .Registers}}
	ioc.AddMockProvider(func() {{$.PkgName}}.{{$register.InterfaceName}} {
		return {{$.PkgName}}mock.{{$register.MethodName}}ByGenerator()
	})
	{{end}}

}
`

const mockConstructorFileTemplate = `
// CODE GENERATED AUTOMATICALLY BY github.com/JianWangEx/commonService/generator/generate_obj_and_mock_file
// THIS FILE SHOULD NOT BE EDITED MANUALLY

var {{$.ObjName}} *{{$.MockName}}
func New{{$.MockName}}ByGenerator() *{{$.MockName}} {
	{{$.ObjName}} = &{{$.MockName}}{}
	return {{$.ObjName}}
}

func Get{{$.MockName}}() *{{$.MockName}} {
	return {{$.ObjName}}
}
`

const registerFileTemplate = `
// CODE GENERATED AUTOMATICALLY BY github.com/JianWangEx/commonService/generator/generate_obj_and_mock_file
// THIS FILE SHOULD NOT BE EDITED MANUALLY

package {{.TypeStr}}

// nolint:funlen
func RegisterProviders() {

}
`
