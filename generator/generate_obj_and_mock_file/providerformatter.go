package generate_obj_and_mock_file

import (
	"bytes"
	"fmt"
	"go/format"
	"sort"
	"strings"
	"text/template"
)

// ProviderOutputFormatter generates the content for providers and mock providers in the memory.
type ProviderOutputFormatter struct {
	Config           Config
	ModuleName       string
	Providers        []ProviderInstance
	GroupedProviders map[string]map[string][]ProviderInstance // same as OutputBytes below
	OutputBytes      map[string]map[string][]byte             // provider.Category=>getProviderFileName(provider)=>content, such as "dao"=>"payment_source_indeximpl_payment_source_dao_impl.go"=>content
	// when IsMock is true, it will contains the constructor, variable/getter for the mock object, which will be appended to the `mockery` generated mock files.
	// key example: "tool/autogenerator/fixtures/spn/handler/mock/spn_handler_mock.go"
	MockConstructorOutputBytes map[string][]byte
	IsMock                     bool
}

func NewProviderOutputFormatter(config Config, registers []ProviderInstance, moduleName string) *ProviderOutputFormatter {
	return &ProviderOutputFormatter{
		Config:                     config,
		ModuleName:                 moduleName,
		Providers:                  registers,
		GroupedProviders:           make(map[string]map[string][]ProviderInstance),
		OutputBytes:                make(map[string]map[string][]byte),
		MockConstructorOutputBytes: make(map[string][]byte),
	}
}

// Format generates the content(in memory) of the providers register for real implementation or mock implementation(If IsMock is true)
//
//	can refer to TestGenerator_FormatOutput_NormalProvider as an example
func (f *ProviderOutputFormatter) Format() error {
	f.groupProvidersByCategoryAndFile()
	tmpl := template.Must(template.New("initFileTemplate").Parse(initFileTemplate))
	if f.IsMock {
		tmpl = template.Must(template.New("mockInitFileTemplate").Parse(mockInitFileTemplate))
	}

	for typeStr, fileToRegisters := range f.GroupedProviders {
		f.OutputBytes[typeStr] = make(map[string][]byte)

		for filename, registers := range fileToRegisters {
			var buf bytes.Buffer
			helper := NewFormatFileHelper(registers)
			err := tmpl.Execute(&buf, helper)
			if err != nil {
				return err
			}
			formatted, err := format.Source(buf.Bytes())
			if err != nil {
				return err
			}
			f.OutputBytes[typeStr][filename] = formatted

			if _, exists := f.OutputBytes[helper.TypeStr]["register.go"]; !exists {
				err := f.getRegisterFileForRegisterType(helper)
				if err != nil {
					return err
				}
			}
		}
	}
	if f.IsMock {
		err := f.formatMockConstructor()
		if err != nil {
			return fmt.Errorf("formatMockConstructor error:%s", err)
		}
	}
	return nil
}

func (f *ProviderOutputFormatter) getRegisterFileForRegisterType(helper *FormatFileHelper) error {
	var buf bytes.Buffer
	tmpl := template.Must(template.New("registerFileTemplate").Parse(registerFileTemplate))
	err := tmpl.Execute(&buf, helper)
	if err != nil {
		return err
	}
	formatted, err := format.Source(buf.Bytes())
	if err != nil {
		return err
	}
	f.OutputBytes[helper.TypeStr]["register.go"] = formatted
	return nil
}

func (f *ProviderOutputFormatter) groupProvidersByCategoryAndFile() {
	for _, reg := range f.Providers {
		filename := getProviderFileName(reg)
		if _, exists := f.GroupedProviders[reg.Category]; !exists {
			f.GroupedProviders[reg.Category] = make(map[string][]ProviderInstance)
		}
		currArr := f.GroupedProviders[reg.Category][filename]
		f.GroupedProviders[reg.Category][filename] = append(currArr, reg)
	}
}

func getProviderFileName(provider ProviderInstance) string {
	return provider.PackageName + "_" + provider.FileName
}

func (f *ProviderOutputFormatter) formatMockConstructor() error {
	tmpl := template.Must(template.New("mockConstructorFileTemplate").Parse(mockConstructorFileTemplate))
	for _, register := range f.Providers {
		fileName := getMockProviderFileName(register, f.ModuleName)
		var buf bytes.Buffer
		helper := NewMockConstructorFileHelper(register)
		err := tmpl.Execute(&buf, helper)
		if err != nil {
			return err
		}
		formatted, err := format.Source(buf.Bytes())
		if err != nil {
			return err
		}
		f.MockConstructorOutputBytes[fileName] = formatted
	}
	return nil
}

func getMockProviderFileName(mockProvider ProviderInstance, moduleName string) string {
	return mockProvider.PackageFullPath[len(moduleName)+1:] + "/" + mockProvider.FileName
}

// Helper structs

type MockConstructorFileHelper struct {
	InterfaceName string
	ObjName       string
	MockName      string
}

func NewMockConstructorFileHelper(instance ProviderInstance) MockConstructorFileHelper {
	helper := MockConstructorFileHelper{}
	helper.InterfaceName = instance.InterfaceName
	helper.MockName = instance.InterfaceName + "Mock"
	helper.ObjName = strings.ToLower(instance.InterfaceName[:1]) + instance.InterfaceName[1:]
	return helper
}

type FormatFileHelper struct {
	TypeStr           string
	PkgFullPath       string
	PkgName           string
	ModuleName        string
	Registers         []ProviderInstance
	InterfaceFullPath string // only for mock
}

func NewFormatFileHelper(registers []ProviderInstance) *FormatFileHelper {
	if len(registers) == 0 {
		return nil
	}
	reg := registers[0]
	typeStr := reg.Category
	pkgFullPath := reg.PackageFullPath
	pkgName := reg.PackageName
	interfaceFullPath := reg.InterfaceFullPath
	moduleName := reg.ModuleName

	// sort providers by package and method name
	sort.SliceStable(registers, func(i, j int) bool {
		if registers[i].PackageName == registers[j].PackageName {
			return registers[i].MethodName < registers[j].MethodName
		}
		return registers[i].PackageName < registers[j].PackageName
	})

	return &FormatFileHelper{TypeStr: typeStr, PkgFullPath: pkgFullPath,
		PkgName: pkgName, ModuleName: moduleName, Registers: registers, InterfaceFullPath: interfaceFullPath}
}

type IndexFileHelper struct {
	PartialImportPath string // full package name of output root directory
	PackageName       string
	TypeArr           []string
}

func NewIndexFileHelper(partialImportPath string, packageName string, typeArr []string) *IndexFileHelper {
	sort.Strings(typeArr)
	return &IndexFileHelper{PartialImportPath: partialImportPath, PackageName: packageName, TypeArr: typeArr}
}
