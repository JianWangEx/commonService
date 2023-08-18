package generate_obj_and_mock_file

// Templates

const injectorFileTemplate = `
// CODE GENERATED AUTOMATICALLY BY github.com/JianWangEx/commonService/generator/generate_obj_and_mock_file
// THIS FILE SHOULD NOT BE EDITED MANUALLY

package {{.PkgName}}

import (
	"{$injector.ModuleName}/pkg/ioc"
)

{{range $injector := .Injectors}}
var {{$injector.ObjName}} {{$injector.InterfaceName}}
// nolint:funlen
func init() {
	ioc.AddServiceInjector(Inject{{$injector.InterfaceName}})
}

func Inject{{$injector.InterfaceName}}(impl {{$injector.InterfaceName}}) {
	{{$injector.ObjName}} = impl
}

func Get{{$injector.InterfaceName}}() {{$injector.InterfaceName}} {
	return {{$injector.ObjName}}
}
{{end}}
`
