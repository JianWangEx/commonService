package generate_obj_and_mock_file

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

func TestInjectorOutputFormatter_Format_And_MockInterface(t *testing.T) {
	injector1 := InjectorInstance{
		Category:         "spn",
		PackageName:      "handler",
		PackageFullPath:  "github.com/JianWangEx/commonService/generator/generate_obj_and_mock_file/fixtures/spn/handler",
		RelativeFilePath: "fixtures/spn/handler/spn_handler.go",
		FileName:         "spn_handler.go",
		InterfaceName:    "SpnHandler",
		ObjName:          "spnHandlerObj",
	}
	injector2 := InjectorInstance{
		Category:         "game",
		PackageName:      "manager",
		PackageFullPath:  "github.com/JianWangEx/commonService/generator/generate_obj_and_mock_file/fixtures/game/manager",
		RelativeFilePath: "fixtures/game/manager/game_manager.go",
		FileName:         "game_manager.go",
		InterfaceName:    "GameManager",
		ObjName:          "gameManagerObj",
	}
	formatter := NewInjectorOutputFormatter(
		configForTest(),
		[]InjectorInstance{injector1, injector2},
		"github.com/JianWangEx/commonService/generator/generate_obj_and_mock_file",
	)
	err := formatter.Format()
	require.NoError(t, err)
	outputBytes := formatter.OutputBytes
	require.True(t, len(outputBytes) == 2, "should have two categories")
	require.True(t, len(outputBytes[getInjectorFileName(injector1)]) > 0)
	require.True(t, len(outputBytes[getInjectorFileName(injector2)]) > 0)
	// note the spaces and newlines, based on the template: injectorFileTemplate
	expectedForInjector1 := `// CODE GENERATED AUTOMATICALLY BY github.com/JianWangEx/commonService/generator/generate_obj_and_mock_file
// THIS FILE SHOULD NOT BE EDITED MANUALLY

package handler

import (
	"github.com/JianWangEx/commonService/pkg/ioc"
)

var spnHandlerObj SpnHandler

// nolint:funlen
func init() {
	ioc.AddServiceInjector(InjectSpnHandler)
}

func InjectSpnHandler(impl SpnHandler) {
	spnHandlerObj = impl
}

func GetSpnHandler() SpnHandler {
	return spnHandlerObj
}
`
	content1 := string(outputBytes[getInjectorFileName(injector1)])
	assert.Equal(t, expectedForInjector1, content1)

	expectedForInjector2 := `// CODE GENERATED AUTOMATICALLY BY github.com/JianWangEx/commonService/generator/generate_obj_and_mock_file
// THIS FILE SHOULD NOT BE EDITED MANUALLY

package manager

import (
	"github.com/JianWangEx/commonService/pkg/ioc"
)

var gameManagerObj GameManager

// nolint:funlen
func init() {
	ioc.AddServiceInjector(InjectGameManager)
}

func InjectGameManager(impl GameManager) {
	gameManagerObj = impl
}

func GetGameManager() GameManager {
	return gameManagerObj
}
`
	content2 := string(outputBytes[getInjectorFileName(injector2)])
	assert.Equal(t, expectedForInjector2, content2)

	// Testify Mock Interface
	// remove the old mock first
	gameManagerMockDir := "fixtures/game/manager/mock"
	spnHandlerMockDir := "fixtures/spn/handler/mock"
	var removeMock = func() {
		os.RemoveAll(gameManagerMockDir)
		os.RemoveAll(spnHandlerMockDir)
	}
	removeMock()
	t.Cleanup(removeMock)
	_, err = os.Stat(gameManagerMockDir)
	require.True(t, os.IsNotExist(err), "should not have game manager mock dir")
	_, err = os.Stat(spnHandlerMockDir)
	require.True(t, os.IsNotExist(err), "should not have spn handler mock dir")
	mockProviders, err := formatter.MockInterface()
	require.NoError(t, err)
	require.Equal(t, 2, len(mockProviders), "each interface has one mock provider")
	_, err = os.Stat(gameManagerMockDir + "/game_manager_mock.go")
	require.NoError(t, err, "should generate game manager mock file")
	_, err = os.Stat(spnHandlerMockDir + "/spn_handler_mock.go")
	require.NoError(t, err, "should generate spn handler mock file")
}
