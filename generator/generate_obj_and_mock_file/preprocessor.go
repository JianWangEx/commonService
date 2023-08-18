package generate_obj_and_mock_file

import (
	"bufio"
	"fmt"
	"github.com/pkg/errors"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

// Preprocessor is responsible for getting all packages in the project
// and establish relationship between package name, path and full path
type Preprocessor struct {
	config             Config
	ModuleName         string            // like github.com/JianWangEx/commonService
	PkgPathToName      map[string]string // like ./generator/generate_obj_and_mock_file/fixtures/spn/manager/impl => spnmanagerimpl
	PkgNameToFullPath  map[string]string // like spnmanagerimpl => github.com/JianWangEx/commonService/generator/generate_obj_and_mock_file/fixtures/spn/manager/impl
	modFilePathForTest string            // it's just for unit test
}

func NewPreprocessor(config Config) *Preprocessor {
	return &Preprocessor{
		config:            config,
		PkgPathToName:     make(map[string]string),
		PkgNameToFullPath: make(map[string]string),
	}
}

// Process scans project to find packages which need to be processed. Basically, it will do these things:
//  1. get all package list by cmd `go list ./...`;
//  2. get module name by reading `go.mod` file;
//  3. walk through all packages and files, find the files under config.ScanPkgs(excluding the mock pkg and files);
//  4. try to create distinct package name for each package for later process.
//
// Return Example:
//
//	pkgPathToName: ./generator/generate_obj_and_mock_file/fixtures/spn/manager/impl => spnmanagerimpl
//	pkgNameToFullPath: spnmanagerimpl => github.com/JianWangEx/commonService/generator/generate_obj_and_mock_file/fixtures/spn/manager/impl
//	moduleName: github.com/JianWangEx/commonService
func (p *Preprocessor) Process() (pkgPathToName map[string]string, pkgNameToFullPath map[string]string, moduleName string, err error) {
	fullPathArr, err := p.getAllPackageFullPath()
	if err != nil {
		return nil, nil, "", err
	}
	p.ModuleName, err = p.getModuleName()
	if err != nil {
		return nil, nil, "", err
	}
	for _, fullPath := range fullPathArr {
		path, err := p.getLocalizedPathFromFullPath(fullPath, p.ModuleName)
		if err != nil {
			return nil, nil, "", err
		}
		skip := true
		for _, scanPkg := range p.config.ScanPkgs {
			if strings.HasPrefix(fullPath, p.ModuleName+scanPkg) {
				skip = false
				break
			}
		}
		if skip {
			continue
		}
		// skip mock
		if strings.Contains(fullPath, "/mock/") || strings.HasSuffix(fullPath, "/mock") {
			continue
		}
		name := p.getPackageNameFromPath(path)
		p.addPackage(fullPath, path, name)
	}
	// why don't we filter during packages scanning? because we need to generate distinct names which needs scanning all files to avoid duplication
	p.Filter()
	return p.PkgPathToName, p.PkgNameToFullPath, p.ModuleName, nil
}

func (p *Preprocessor) Filter() {
	keyword := p.config.FilterKeyword
	if len(keyword) == 0 {
		return
	}
	var pkgNameToBeMoved []string
	for pkgPath, pkgName := range p.PkgPathToName {
		// ./generator/generate_obj_and_mock_file/fixtures/spn/manager/impl => spnmanagerimpl
		// need append the "/", otherwise cannot match the root dir, for example, "./internal/services/course" need to match the filter "/course/"
		if !strings.Contains(pkgPath+"/", keyword) {
			delete(p.PkgPathToName, pkgPath)
			pkgNameToBeMoved = append(pkgNameToBeMoved, pkgName)
		}
	}
	for pkgName, pkgFullPath := range p.PkgNameToFullPath {
		// spnmanagerimpl => github.com/JianWangEx/commonService/generator/generate_obj_and_mock_file/fixtures/spn/manager/impl
		if InSlice(pkgName, pkgNameToBeMoved) && !strings.Contains(pkgFullPath+"/", keyword) {
			delete(p.PkgNameToFullPath, pkgName)
		}
	}
}

func (p *Preprocessor) addPackage(fullPath, path, name string) {
	nonConflictingName := p.getNonConflictingName(path, name)
	p.PkgPathToName[path] = nonConflictingName
	p.PkgNameToFullPath[nonConflictingName] = fullPath
}

func (p *Preprocessor) getNonConflictingName(path, name string) string {
	// Combine parent package name to the given name
	directories := strings.Split(path, "/")
	var invalidIdentifierChar = regexp.MustCompile("[^[:digit:][:alpha:]_]")
	cleanedDirectories := make([]string, 0, len(directories))
	for _, directory := range directories {
		cleaned := invalidIdentifierChar.ReplaceAllString(directory, "_")
		cleanedDirectories = append(cleanedDirectories, cleaned)
	}
	numDirectories := len(cleanedDirectories)
	var prospectiveName string
	for i := 2; i <= numDirectories; i++ {
		prospectiveName = strings.Join(cleanedDirectories[numDirectories-i:], "")
		if !p.importNameExists(prospectiveName) {
			return prospectiveName
		}
	}
	// Try adding numbers to the given name if there are still conflicts
	i := 2
	for {
		prospectiveName = fmt.Sprintf("%v%d", name, i)
		if !p.importNameExists(prospectiveName) {
			return prospectiveName
		}
		i++
	}
}

func (p *Preprocessor) importNameExists(name string) bool {
	_, nameExists := p.PkgNameToFullPath[name]
	return nameExists
}

func (p *Preprocessor) getLocalizedPathFromFullPath(fullPath string, rootPath string) (string, error) {
	if fullPath != "" && !strings.Contains(fullPath, rootPath) {
		return "", &ReadModuleNameError{}
	}
	return strings.Replace(fullPath, rootPath, ".", -1), nil
}

func (p *Preprocessor) getModuleName() (string, error) {
	goModFile := "go.mod"
	if p.modFilePathForTest != "" {
		goModFile = p.modFilePathForTest
	}
	file, err := os.Open(goModFile)
	if err != nil {
		return "", err
	}

	scanner := bufio.NewScanner(file)
	scanner.Scan()
	replacer := strings.NewReplacer("module ", "")
	line := scanner.Text()
	moduleName := replacer.Replace(line)

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
	err = file.Close()
	if err != nil {
		fmt.Println(err.Error())
		return "", err
	}
	return moduleName, nil
}

func (p *Preprocessor) getPackageNameFromPath(path string) string {
	arr := strings.Split(path, "/")
	return arr[len(arr)-1]
}

func (p *Preprocessor) getAllPackageFullPath() ([]string, error) {
	// Get all packages in the current project
	data, err := exec.Command("bash", "-c", getAllPackagesCmd).CombinedOutput()
	if err != nil {
		return nil, errors.WithMessagef(err, "Run command '%s' error:\n%s", getAllPackagesCmd, string(data))
	}

	fullPathArr := strings.Split(string(data), "\n")
	if len(fullPathArr) == 0 {
		return nil, err
	}
	return fullPathArr, nil
}

type ReadModuleNameError struct {
}

func (e *ReadModuleNameError) Error() string {
	return "Error reading module name. Make sure the project has a go.mod file with module name specified."
}

func InSlice[T comparable](key T, sets []T) bool {
	for _, k := range sets {
		if key == k {
			return true
		}
	}
	return false
}
