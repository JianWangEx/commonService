package generate_obj_and_mock_file

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"
)

// ProviderInstance represents a constructor instance found used in fx as a Provider to the Injector
//
//	Can refer to TestParser_Parse as an example
type ProviderInstance struct {
	Category          string // one of the pkg names of config.ScanPkgs, such as "dao", "business"
	PackageName       string // distinct pkg name, like "handlerimpl"
	PackageFullPath   string // like "github.com/JianWangEx/commonService/generator/generate_obj_and_mock_file/fixtures/spn/handler/impl"
	RelativeFilePath  string // relative to the root directory of the go module, like "tool/autogenerator/fixtures/spn/handler/impl/spn_handler_impl.go"
	FileName          string // like "spn_handler_impl.go"
	MethodName        string // like "NewSpnHandlerImpl", normal function name of the constructor you defined.
	InterfaceFullPath string // only for mock file format step, like "github.com/JianWangEx/commonService/generator/generate_obj_and_mock_file/fixtures/spn/handler"
	InterfaceName     string // only for mock file format step
}

// InjectorInstance represents an injector instance found used in fx as Invoker params, which will call this to inject the concrete implementation.
//
//	Can refer to TestParser_Parse as an example
type InjectorInstance struct {
	Category         string // one of the pkg names of config.ScanPkgs, such as "dao", "service"
	PackageName      string // distinct pgk name, like "handler", "daocourse"
	PackageFullPath  string // like "github.com/JianWangEx/commonService/internal/data/dao"
	RelativeFilePath string // relative to the root directory of the go module, like "internal/data/dao/gym_dao.go"
	FileName         string // like "gym_dao.go"
	InterfaceName    string // like "GymDao", normal interface name as you defined.
	ObjName          string // like "gymDaoObj", it's used later as global variable name which will hold the concrete implementation injected by fx
}

// Generator is responsible for coordinating the overall flow
// of generating the file which initializes all providers in the project
type Generator struct {
	Config                     Config
	ModuleName                 string            // Example: github.com/JianWangEx/commonService
	PkgPathToName              map[string]string // Example: ./tool/autogenerator/fixtures/spn/manager/impl => spnmanagerimpl
	PkgNameToFullPath          map[string]string // Example: spnmanagerimpl => github.com/JianWangEx/commonService/generator/generate_obj_and_mock_file/fixtures/spn/manager/impl
	providers                  []ProviderInstance
	injectors                  []InjectorInstance
	ProviderOutputBytes        map[string]map[string][]byte
	MockProviderOutputBytes    map[string]map[string][]byte
	MockConstructorOutputBytes map[string][]byte
	InjectorsOutputBytes       map[string][]byte
}

// NewGenerator builds a Generator.
func NewGenerator(config Config) *Generator {
	return &Generator{
		Config:               config,
		PkgPathToName:        make(map[string]string),
		PkgNameToFullPath:    make(map[string]string),
		ProviderOutputBytes:  make(map[string]map[string][]byte),
		InjectorsOutputBytes: make(map[string][]byte),
	}
}

func (g *Generator) Generate() (err error) {
	if g.Config.FilterKeyword == "" {
		fmt.Fprint(os.Stdout, colorize("!!!Recommendation: Use partial model like 'make ag filter=/payment/' to make it faster.\nThe 'filter' can be any part of the path. Such as 'inbound', 'thirdparty', 'lianlian' .etc.\n", colorMagenta))
	}
	startTime := time.Now()
	// Step 1. scans project to find packages which need to be processed, details can refer to Preprocessor.Process
	fmt.Fprintf(os.Stdout, "%s%s: Scanning Project, PartialModel=%t, Filter=%s, ScanPkgs=%s\n",
		timeStr(), colorize("Preprocess", colorCyan), g.Config.PartialMode(), g.Config.FilterKeyword, g.Config.ScanPkgs)
	err = g.Preprocess()
	if err != nil {
		fmt.Fprint(os.Stderr, colorize(fmt.Sprintf("Error while preprocessing: %v\n", err), colorRed))
		return
	}
	if g.Config.FilterKeyword != "" && g.Config.Debug {
		fmt.Fprint(os.Stdout, colorize("\tAfter Filter, Actual packages:\n", colorDarkGray))
		count := 0
		for pkgPath := range g.PkgPathToName {
			count++
			fmt.Fprint(os.Stdout, colorize(fmt.Sprintf("\t(%d/%d)%s\n", count, len(g.PkgPathToName), pkgPath), colorDarkGray))
		}
	}
	// Step 2. parses pkgs returns from step 1, collects the eligible interfaces as []InjectorInstance and constructors as []ProviderInstance
	fmt.Fprintf(os.Stdout, "%s%s: finding the providers and injectors\n", timeStr(), colorize("Parsing Files", colorCyan))
	err = g.Parse()
	if err != nil {
		fmt.Fprint(os.Stderr, colorize(fmt.Sprintf("Error while parsing files: %v\n", err), colorRed))
		return
	}
	// Step 3. Build content for providers and injectors. Most operations are in memory, except the mock file generation for interface
	fmt.Fprintf(os.Stdout, "%s%s: generate file content and call `mockery` to create mock files(Providers=%d, Injector=%d)\n",
		timeStr(), colorize("Building Output", colorCyan), len(g.providers), len(g.injectors))
	err = g.FormatOutputAndGenerateMockFiles()
	if err != nil {
		fmt.Fprint(os.Stderr, colorize(fmt.Sprintf("Error while formatting output: %v\n", err), colorRed))
		return
	}
	// Step 4. The real write happens here(expect mock file which happens in step 3)
	fmt.Fprintf(os.Stdout, "%s%s: Create Injector Files and Register Files For Providers\n", timeStr(), colorize("Writing files", colorCyan))
	err = g.Write()
	if err != nil {
		fmt.Fprint(os.Stderr, colorize(fmt.Sprintf("Error while writing to file: %v\n", err), colorRed))
		return
	}
	fmt.Fprintf(os.Stdout, "%s%s\n", timeStr(), colorize(fmt.Sprintf("Generate Finished, Cost=%s", time.Since(startTime)), colorGreen))
	return nil
}

// Write writes contents to files
func (g *Generator) Write() error {
	// Create and write register files for provider(ioc.AddProvider...) to Config.OutputDirPathName(such as dir "internal/register")
	providerWriter := NewProviderWriter(g.Config, g.ProviderOutputBytes)
	err := providerWriter.Write()
	if err != nil {
		return err
	}
	// Create and write injector files besides the interface file.
	injectorWriter := NewInjectorWriter(g.Config, g.InjectorsOutputBytes)
	err = injectorWriter.Write()
	if err != nil {
		return err
	}

	// Create and write register files for mock provider(ioc.AddMockProvider...) to Config.OutputDirPathName(such as dir "internal/register/mock")
	mockProviderWriter := NewProviderWriter(g.Config, g.MockProviderOutputBytes)
	mockProviderWriter.IsMock = true
	err = mockProviderWriter.Write()
	if err != nil {
		return err
	}

	// append convenient helpers like getter/constructor for the mock object to `mockery` generated mock file.
	mockConstructorWriter := NewInjectorWriter(g.Config, g.MockConstructorOutputBytes)
	mockConstructorWriter.IsMockAppend = true
	err = mockConstructorWriter.Write()
	if err != nil {
		return err
	}
	g.WriteCheckSum()
	return nil
}

func (g *Generator) WriteCheckSum() {
	if !g.Config.Incremental {
		return
	}
	filePath := filepath.Join(g.Config.OutputDirPathName, g.Config.CheckSumFileName)
	f, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		fmt.Printf("create %s error: %+v\n", filePath, err)
		return
	}
	defer f.Close()
	var fileList []string
	for _, register := range g.providers {
		fileList = append(fileList, register.RelativeFilePath)
	}
	for _, injector := range g.injectors {
		fileList = append(fileList, injector.RelativeFilePath)
	}
	sort.Strings(fileList)
	for _, relativeFilePath := range fileList {
		fileList = append(fileList, relativeFilePath)
		stat, err := os.Stat(relativeFilePath)
		if err != nil {
			fmt.Printf("os.Stat%s) error:%+v", relativeFilePath, err)
			continue
		}
		_, err = f.WriteString(fmt.Sprintf("%s:%d:%d\n", relativeFilePath, stat.ModTime().Unix(), stat.Size()))
		if err != nil {
			fmt.Printf("WriteString to file(%s) error:%+v", relativeFilePath, err)
			continue
		}
	}
}

// FormatOutputAndGenerateMockFiles generates content in memory for providers and injectors, as well as running "mockery" cmd to generate mock files.
func (g *Generator) FormatOutputAndGenerateMockFiles() error {
	// Generates the content(in memory) of the providers register for real implementation
	providerFormatter := NewProviderOutputFormatter(g.Config, g.providers, g.ModuleName)
	err := providerFormatter.Format()
	if err != nil {
		return err
	}
	g.ProviderOutputBytes = providerFormatter.OutputBytes

	// Generates the content(in memory) of the injectors for interface
	injectorFormatter := NewInjectorOutputFormatter(g.Config, g.injectors, g.ModuleName)
	err = injectorFormatter.Format()
	if err != nil {
		return err
	}
	g.InjectorsOutputBytes = injectorFormatter.OutputBytes
	// Run "mockery" cmd to generate mock files for the interface(it will write new files)
	mockRegisters, err := injectorFormatter.MockInterface()
	if err != nil {
		return err
	}

	// Generates the content(in memory) of the providers register for mock implementation
	// as well as some helper methods for the mock objects like getter/constructor,
	// which will be appended to the `mockery` generated mock file.
	mockFormatter := NewProviderOutputFormatter(g.Config, mockRegisters, g.ModuleName)
	mockFormatter.IsMock = true
	err = mockFormatter.Format()
	if err != nil {
		return err
	}
	g.MockProviderOutputBytes = mockFormatter.OutputBytes                   // mock provider register content
	g.MockConstructorOutputBytes = mockFormatter.MockConstructorOutputBytes // append content to the mock file
	return nil
}

func (g *Generator) Parse() error {
	parser := NewParser(g.Config)
	for path, name := range g.PkgPathToName {
		if path != "" {
			err := parser.Parse(path, name, g.PkgNameToFullPath[name])
			if err != nil {
				return err
			}
		}
	}
	g.providers = parser.providers
	g.injectors = parser.injectors
	return nil
}

func (g *Generator) Preprocess() error {
	processor := NewPreprocessor(g.Config)
	var err error
	g.PkgPathToName, g.PkgNameToFullPath, g.ModuleName, err = processor.Process()
	return err
}

func timeStr() string {
	return colorize(time.Now().Format("[15:04:05]"), colorDarkGray)
}
