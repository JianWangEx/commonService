package generate_obj_and_mock_file

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestParser_Parse(t *testing.T) {
	config := configForTest()
	config.ScanPkgs = []string{"/fixtures/spn/"} // need to do some trick because of the relative directory in unit test
	parser := NewParser(config)
	pkgPathToName := map[string]string{
		"./fixtures/spn/handler":      "spnhandler",  // interface, aka injector
		"./fixtures/spn/handler/impl": "handlerimpl", // concrete implementation, aka provider
	}
	pkgNameToFullPath := map[string]string{
		"spnhandler":  "github.com/JianWangEx/commonService/generator/generate_obj_and_mock_file/fixtures/spn/handler",
		"handlerimpl": "github.com/JianWangEx/commonService/generator/generate_obj_and_mock_file/fixtures/spn/handler/impl",
	}
	for path, name := range pkgPathToName {
		err := parser.Parse(path, name, pkgNameToFullPath[name])
		assert.NoErrorf(t, err, "should not have error when parse:%s, %s", path, name)
	}
	injectors := parser.injectors
	providers := parser.providers
	require.Equal(t, 1, len(injectors), "one injector")
	require.Equal(t, 1, len(providers), "one provider")
	expectedInjector := InjectorInstance{
		Category:         "spn",
		PackageName:      "handler",
		PackageFullPath:  "github.com/JianWangEx/commonService/generator/generate_obj_and_mock_file/fixtures/spn/handler",
		RelativeFilePath: "fixtures/spn/handler/spn_handler.go",
		FileName:         "spn_handler.go",
		InterfaceName:    "SpnHandler",
		ObjName:          "spnHandlerObj",
	}
	expectedProvider := ProviderInstance{
		Category:          "spn",
		PackageName:       "handlerimpl",
		PackageFullPath:   "github.com/JianWangEx/commonService/generator/generate_obj_and_mock_file/fixtures/spn/handler/impl",
		RelativeFilePath:  "fixtures/spn/handler/impl/spn_handler_impl.go",
		FileName:          "spn_handler_impl.go",
		MethodName:        "NewSpnHandlerImpl",
		InterfaceFullPath: "",
		InterfaceName:     "",
	}
	assert.Equal(t, expectedInjector, injectors[0], "injector instance")
	assert.Equal(t, expectedProvider, providers[0], "provider instance")
}

func TestParser_getCategory(t *testing.T) {
	config := configForTest()
	tests := []struct {
		commentStr string
		fileName   string
		identifier string
		want       string
		msg        string
	}{
		{
			commentStr: "// @Injector()", fileName: "tool/autogenerator/fixtures/spn/handler/spn_handler.go", identifier: DefaultConfig.IdentifierOfInjector,
			want: "spn", msg: "the very normal one for injector, should be the last fragment of the matched scan pkg with file name",
		},
		{
			commentStr: "// more before @Injector() more after", fileName: "tool/autogenerator/fixtures/spn/handler/spn_handler.go", identifier: DefaultConfig.IdentifierOfInjector,
			want: "spn", msg: "should contain any additional comments",
		},
		{
			commentStr: "// @Injector()", fileName: "/tool/autogenerator/fixtures/spn/handler/spn_handler.go", identifier: DefaultConfig.IdentifierOfInjector,
			want: "", msg: "cannot support '/' at the first",
		},
		{
			commentStr: "// @Injector()", fileName: "nowhere/spn_handler.go", identifier: DefaultConfig.IdentifierOfInjector,
			want: "", msg: "file name not under any of the config.ScanPkgs",
		},
		{
			commentStr: "// @Injector", fileName: "nowhere/spn_handler.go", identifier: DefaultConfig.IdentifierOfInjector,
			want: "", msg: "must have () after @Injector)",
		},
		{
			commentStr: "//@Provider()", fileName: "tool/autogenerator/fixtures/spn/handler/spn_handler.go", identifier: DefaultConfig.IdentifierOfProvider,
			want: "spn", msg: "the very normal one for provider, should be the last fragment of the matched scan pkg with file name",
		},
		{
			commentStr: "// @Provider()", fileName: "/tool/autogenerator/fixtures/spn/handler/spn_handler.go", identifier: DefaultConfig.IdentifierOfProvider,
			want: "", msg: "cannot support '/' at the first",
		},
	}
	for _, tt := range tests {
		name := fmt.Sprintf("(%v, %v, %v)", tt.commentStr, tt.fileName, tt.identifier)
		t.Run(name, func(t *testing.T) {
			p := NewParser(config)
			assert.Equalf(t, tt.want, p.getCategory(tt.commentStr, tt.fileName, tt.identifier), fmt.Sprintf("%s. Scan pkgs: %s", tt.msg, config.ScanPkgs))
		})
	}
}
