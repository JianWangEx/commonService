package generate_obj_and_mock_file

import (
	"bytes"
	"fmt"
	"go/format"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"text/template"
	"time"
)

// InjectorOutputFormatter generates the content for injectors for interfaces.
//
//	It's very similar with ProviderOutputFormatter, except that it will run "mockery" to generate mock files.
type InjectorOutputFormatter struct {
	Config     Config
	ModuleName string
	Injectors  []InjectorInstance
	// The first level group(category) is redundant, see OutputBytes below. Second level is the relative xxx_injector.go file name.
	// Btw, at first I want to put this files together under one package as providers do,
	//	But different interfaces depend on models in different package, if you put these injectors together,
	//	it means import all business models at once, which makes it easy to cause cycle import. Besides, the names of methods are difficult to keep both distinct and elegant to use.
	GroupedInjectors map[string]map[string][]InjectorInstance
	// Key example: "tool/autogenerator/fixtures/spn/handler/spn_handler_injector.go",
	//	Unlike provider, the injector file is put together with the interface file, so no need to group by category
	OutputBytes map[string][]byte
}

func NewInjectorOutputFormatter(config Config, injectors []InjectorInstance, moduleName string) *InjectorOutputFormatter {
	return &InjectorOutputFormatter{
		Config:           config,
		ModuleName:       moduleName,
		Injectors:        injectors,
		GroupedInjectors: make(map[string]map[string][]InjectorInstance),
		OutputBytes:      make(map[string][]byte),
	}
}

// Format generates the content for injectors in memory
func (f *InjectorOutputFormatter) Format() error {
	f.groupInjectorsByCategoryAndFile()
	tmpl := template.Must(template.New("injectorFileTemplate").Parse(injectorFileTemplate))
	for _, fileToRegisters := range f.GroupedInjectors {
		for filename, registers := range fileToRegisters {
			var buf bytes.Buffer
			helper := NewInjectorFormatFileHelper(registers)
			err := tmpl.Execute(&buf, helper)
			if err != nil {
				return err
			}
			formatted, err := format.Source(buf.Bytes())
			if err != nil {
				return err
			}
			f.OutputBytes[filename] = formatted
		}
	}
	return nil
}

const mockeryTemplate = "mockery --quiet=false --name=%s --case=underscore --structname=%s --filename=%s --with-expecter=true --dir=%s --output=%s --outpkg=%s"

// MockInterface runs "mockery" to generate mock files for interface, returns the provider instance for the mock implementation.
func (f *InjectorOutputFormatter) MockInterface() ([]ProviderInstance, error) {
	var mockRegisters []ProviderInstance
	var mockeryCmdList []string
	checksums, err := f.GetCheckSum()
	if err != nil {
		fmt.Println("GetCheckSum error", err)
	}
	for _, injector := range f.Injectors {
		name := injector.InterfaceName                                         // like ClientApiPermissionDao
		structName := name + "Mock"                                            // append "Mock" to interface name, like ClientApiPermissionDaoMock
		fileName := injector.FileName[0:len(injector.FileName)-3] + "_mock.go" // append "_mock" to interface file name, like client_api_permission_dao_mock.go
		dir := injector.PackageFullPath[len(f.ModuleName)+1:]                  // the relative dir relative to module name, like "internal/data/dao/client_api_permission"
		output := dir + "/mock"                                                // like "internal/data/dao/client_api_permission/mock"
		outpkg := injector.PackageName + "mock"                                // like "client_api_permissionmock"
		if lastChangedTime, ok := checksums[injector.RelativeFilePath]; ok {
			stat, err := os.Stat(injector.RelativeFilePath)
			if err != nil {
				fmt.Println("os.Stat error", injector.RelativeFilePath)
			} else if stat.ModTime().Unix() <= lastChangedTime {
				// check if the mock file exist, only skip when it's already exist, or we still need to generate
				fullMockFileName := filepath.Join(output, fileName)
				if _, err = os.Stat(fullMockFileName); err == nil {
					continue
				}
			}
		}
		mockeryCmd := fmt.Sprintf(mockeryTemplate, name, structName, fileName, dir, output, outpkg)
		mockeryCmdList = append(mockeryCmdList, mockeryCmd)
		instance := ProviderInstance{
			Category:          injector.Category,
			PackageName:       injector.PackageName,
			PackageFullPath:   injector.PackageFullPath + "/mock",
			FileName:          fileName,
			MethodName:        "New" + injector.InterfaceName + "Mock",
			InterfaceFullPath: injector.PackageFullPath,
			InterfaceName:     injector.InterfaceName,
		}
		mockRegisters = append(mockRegisters, instance)
	}
	if len(mockRegisters) > 0 {
		runMockery(mockeryCmdList)
	}
	return mockRegisters, nil
}

func (f *InjectorOutputFormatter) GetCheckSum() (map[string]int64, error) {
	if !f.Config.Incremental {
		return nil, nil
	}
	checkSumFile := filepath.Join(f.Config.OutputDirPathName, f.Config.CheckSumFileName)
	content, err := ioutil.ReadFile(checkSumFile)
	if err != nil {
		return nil, err
	}
	if len(content) == 0 {
		return nil, nil
	}
	res := map[string]int64{}
	checkSums := strings.Split(string(content), "\n")
	for _, checkSum := range checkSums {
		fragments := strings.Split(checkSum, ":")
		if len(fragments) < 2 {
			continue
		}
		lastChangedTime, err := strconv.Atoi(fragments[1])
		if err != nil {
			fmt.Println("read checksum last changed time error", checkSums, err)
			continue
		}
		res[fragments[0]] = int64(lastChangedTime)
	}
	return res, nil
}

func runMockery(mockeryCmdList []string) {
	batchSize := 3
	concurrency := os.Getenv("ag_concurrency")
	if concurrency != "" {
		concurrencyNum, err := strconv.ParseInt(concurrency, 10, 64)
		if err != nil {
			log.Printf("invalid concurrent env:%s, err:%s", concurrency, err)
		} else {
			batchSize = int(concurrencyNum)
		}
	}
	startTime := time.Now()
	fmt.Print(colorize(fmt.Sprintf("Generating mock files, total=%d, concurrency=%d\n", len(mockeryCmdList), batchSize), colorDarkGray))
	batchList := map[int][]string{}
	for i := range mockeryCmdList {
		batch := i % batchSize
		cmds := batchList[batch]
		batchList[batch] = append(cmds, mockeryCmdList[i])
	}
	waitGroup := sync.WaitGroup{}
	spin := NewChin(SpinSet(36, colorYellow))
	go spin.Start()
	for _, batchCmd := range batchList {
		waitGroup.Add(1)
		go func(cmdList []string) {
			defer waitGroup.Done()
			for _, cmd := range cmdList {
				command := exec.Command("bash", "-c", cmd)
				result, err := command.CombinedOutput()
				if err != nil {
					log.Fatal(colorize(fmt.Sprintf("Generating mock files error, cmd=%s, err=%s\n, msg=%s", cmd, err, string(result)), colorRed))
				}
			}
		}(batchCmd)
	}
	waitGroup.Wait()
	spin.Stop()
	fmt.Fprint(os.Stdout, colorize(fmt.Sprintf("Generating mock files end, total=%d, concurrency=%d, cost=%s\n", len(mockeryCmdList), batchSize, time.Since(startTime)), colorDarkGray))
}

func (f *InjectorOutputFormatter) groupInjectorsByCategoryAndFile() {
	for _, reg := range f.Injectors {
		filename := getInjectorFileName(reg)
		if _, exists := f.GroupedInjectors[reg.Category]; !exists {
			f.GroupedInjectors[reg.Category] = make(map[string][]InjectorInstance)
		}
		currArr := f.GroupedInjectors[reg.Category][filename]
		f.GroupedInjectors[reg.Category][filename] = append(currArr, reg)
	}
}

func getInjectorFileName(injector InjectorInstance) string {
	return injector.RelativeFilePath[:len(injector.RelativeFilePath)-3] + "_injector.go"
}

// Helper structs

type InjectorFormatFileHelper struct {
	TypeStr     string
	PkgFullPath string
	PkgName     string
	Injectors   []InjectorInstance
}

func NewInjectorFormatFileHelper(injectors []InjectorInstance) *InjectorFormatFileHelper {
	if len(injectors) == 0 {
		return nil
	}
	reg := injectors[0]
	typeStr := reg.Category
	pkgFullPath := reg.PackageFullPath
	pkgName := reg.PackageName

	// sort providers by package and method name
	sort.SliceStable(injectors, func(i, j int) bool {
		if injectors[i].PackageName == injectors[j].PackageName {
			return injectors[i].InterfaceName < injectors[j].InterfaceName
		}
		return injectors[i].PackageName < injectors[j].PackageName
	})
	return &InjectorFormatFileHelper{TypeStr: typeStr, PkgFullPath: pkgFullPath,
		PkgName: pkgName, Injectors: injectors}
}
