package generate_obj_and_mock_file

import (
	"bufio"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"os"
	"regexp"
	"strings"
	"unicode"
)

// Parser is responsible for parsing the project code base, look for function
// declaration with the comment identifier for providers and parse them to an
// array of ProviderInstance
type Parser struct {
	config    Config
	providers []ProviderInstance
	injectors []InjectorInstance
}

func NewParser(config Config) *Parser {
	return &Parser{config: config}
}

// Parse parses the package, find the eligible interfaces and constructors(have specified comment words: @Provider() or @Injector())
//
//	The information of found providers and injectors will be hold by the struct ProviderInstance and InjectorInstance
//	[Params Example]
//	moduleName: github.com/JianWangEx/commonService
//	packagePath: "./fixtures/spn/handler/impl"
//	packageName: "handlerimpl"
//	fullPath: "github.com/JianWangEx/commonService/generator/generate_obj_and_mock_file/fixtures/spn/handler/impl"
func (p *Parser) Parse(moduleName, packagePath, packageName, fullPath string) error {
	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, packagePath, nil, parser.ParseComments)
	if err != nil {
		fmt.Println(err.Error())
		return err
	}
	for _, pkg := range pkgs {
		for fileName, file := range pkg.Files {
			// Inspect the AST
			ast.Inspect(file, func(x ast.Node) bool {
				fn, ok := x.(*ast.FuncDecl)
				if ok {
					_ = p.findRegisterKeywordInComment(moduleName, packageName, fileName, fullPath, fn)
					// Move on to the next sibling node
					return false
				}
				iface, ok := x.(*ast.GenDecl)
				if ok {
					_ = p.findInvokerKeywordInComment(moduleName, fileName, fullPath, iface)
					// Move on to the next sibling node
					return false
				}
				// Invoke function recursively for each of the non-nil children nodes
				return true
			})
		}
	}
	return nil
}

func (p *Parser) findInvokerKeywordInComment(moduleName, fileName, fullPath string, iface *ast.GenDecl) error {
	// Register function must have at least one line of comment
	// It must be exported and interface definition
	if iface.Doc == nil || len(iface.Doc.List) == 0 || len(iface.Specs) != 1 {
		return nil
	}
	spec, ok := iface.Specs[0].(*ast.TypeSpec)
	if !ok {
		return nil
	}
	if !unicode.IsUpper(rune(spec.Name.Name[0])) {
		return nil
	}
	_, ok = spec.Type.(*ast.InterfaceType)
	if !ok {
		return nil
	}

	ifaceName := spec.Name.Name
	for _, comment := range iface.Doc.List {
		injectorCategory := p.getInjectorCategory(comment.Text, fileName)
		if injectorCategory == "" {
			continue
		}
		packageName, err := p.GetPackageNameFromFile(fileName)
		if err != nil {
			fmt.Printf("cannot find package name from file: %v!!\n", fileName)
			return err
		}

		fileNameStr := p.getFileNameStr(fileName)
		objName := strings.ToLower(ifaceName[0:1]) + ifaceName[1:] + "Obj"
		instance := InjectorInstance{
			ModuleName:       moduleName,
			Category:         injectorCategory,
			PackageName:      packageName,
			PackageFullPath:  fullPath,
			RelativeFilePath: fileName,
			FileName:         fileNameStr,
			InterfaceName:    ifaceName,
			ObjName:          objName,
		}

		p.injectors = append(p.injectors, instance)
	}
	return nil
}

func (p *Parser) GetPackageNameFromFile(fileName string) (string, error) {
	file, err := os.Open(fileName)
	if err != nil {
		return "", err
	}

	scanner := bufio.NewScanner(file)
	scanner.Scan()
	replacer := strings.NewReplacer("package ", "")
	line := scanner.Text()
	packageName := replacer.Replace(line)

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
	err = file.Close()
	if err != nil {
		fmt.Println(err.Error())
		return "", err
	}
	return packageName, nil
}

func (p *Parser) findRegisterKeywordInComment(moduleName, packageName, fileName, fullPath string, fn *ast.FuncDecl) error {
	// Register function must have at least one line of comment
	// It has no function receiver, and must be exported
	if fn.Doc == nil || len(fn.Doc.List) == 0 || fn.Recv != nil ||
		!unicode.IsUpper(rune(fn.Name.Name[0])) {
		return nil
	}
	fnName := fn.Name.Name
	for _, comment := range fn.Doc.List {
		providerCategory := p.getProviderCategory(comment.Text, fileName)
		if providerCategory == "" {
			continue
		}
		fileNameStr := p.getFileNameStr(fileName)
		instance := ProviderInstance{
			ModuleName:       moduleName,
			Category:         providerCategory,
			PackageName:      packageName,
			PackageFullPath:  fullPath,
			RelativeFilePath: fileName,
			FileName:         fileNameStr,
			MethodName:       fnName,
		}

		p.providers = append(p.providers, instance)
	}
	return nil
}

func (p *Parser) getProviderCategory(commentStr string, fileName string) string {
	return p.getCategory(commentStr, fileName, p.config.IdentifierOfProvider)
}

func (p *Parser) getInjectorCategory(commentStr string, fileName string) string {
	return p.getCategory(commentStr, fileName, p.config.IdentifierOfInjector)
}

// getCategory try to figure out whether the comment contains the specific identifier and if yes, which pkg of the config.ScanPkgs the file belong to.
//
//	the return category should be one of the last fragments of the config.ScanPkgs.
//	For example, the scanPkgs is ["/internal/dao/", "/internal/business/"], then the category may be "dao" or "business"
func (p *Parser) getCategory(commentStr string, fileName string, identifier string) string {
	// commentStr example: "// @Injector()"
	re := regexp.MustCompile(fmt.Sprintf(`%s\(([a-zA-Z0-9]*)\)`, identifier))
	match := re.FindStringSubmatch(commentStr)
	if len(match) < 1 {
		return ""
	}
	for _, scanPkg := range p.config.ScanPkgs {
		// it's error-prone here, the slash matters a lot, scanPkg example : "/fixtures/spn/"
		// file name example: fixtures/spn/handler/spn_handler.go
		if strings.HasPrefix(fileName, scanPkg[1:]) { // remove the first '/'
			// split example: ["", "fixtures", "spn", ""]
			split := strings.Split(scanPkg, "/")
			// return the last fragment of the scan pkg path, note that the last element of split if empty string, so here need to minus 2
			return split[len(split)-2]
		}
	}
	return ""
}

func (p *Parser) getFileNameStr(fileName string) string {
	arr := strings.Split(fileName, "/")
	fileName = arr[len(arr)-1]
	return fileName
}
