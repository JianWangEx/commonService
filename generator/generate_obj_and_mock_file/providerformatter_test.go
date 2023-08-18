package generate_obj_and_mock_file

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestProviderOutputFormatter_Format_RealImplementationProvider(t *testing.T) {
	// means not provider for mock implementation
	provider1 := ProviderInstance{
		Category:          "spn",
		PackageName:       "handlerimpl",
		PackageFullPath:   "github.com/JianWangEx/commonService/generator/generate_obj_and_mock_file/fixtures/spn/handler/impl",
		RelativeFilePath:  "fixtures/spn/handler/impl/spn_handler_impl.go",
		FileName:          "spn_handler_impl.go",
		MethodName:        "NewSpnHandlerImpl",
		InterfaceFullPath: "",
		InterfaceName:     "",
	}
	provider2 := ProviderInstance{
		Category:          "game",
		PackageName:       "managerimpl",
		PackageFullPath:   "github.com/JianWangEx/commonService/generator/generate_obj_and_mock_file/fixtures/game/manager/impl",
		RelativeFilePath:  "fixtures/game/manager/impl/game_manager_impl.go",
		FileName:          "game_manager_impl.go",
		MethodName:        "NewGameManagerImpl",
		InterfaceFullPath: "",
		InterfaceName:     "",
	}
	formatter := NewProviderOutputFormatter(configForTest(), []ProviderInstance{provider1, provider2}, "github.com/JianWangEx/commonService")
	err := formatter.Format()
	require.NoError(t, err)
	outputBytes := formatter.OutputBytes
	require.True(t, len(outputBytes) == 2, "should have two categories")
	require.True(t, len(outputBytes[provider1.Category]) > 0)
	require.True(t, len(outputBytes[provider2.Category]) > 0)
	// note the spaces and newlines, based on the template: initFileTemplate
	expectedForProvider1 := `// CODE GENERATED AUTOMATICALLY BY github.com/JianWangEx/commonService/generator/generate_obj_and_mock_file
// THIS FILE SHOULD NOT BE EDITED MANUALLY

package spn

import (
	handlerimpl "github.com/JianWangEx/commonService/generator/generate_obj_and_mock_file/fixtures/spn/handler/impl"
	"github.com/JianWangEx/commonService/pkg/ioc"
)

// nolint:funlen
func init() {

	ioc.AddProvider(handlerimpl.NewSpnHandlerImpl)

}
`
	content1 := string(outputBytes[provider1.Category][getProviderFileName(provider1)])
	assert.Equal(t, expectedForProvider1, content1)

	expectedForProvider2 := `// CODE GENERATED AUTOMATICALLY BY github.com/JianWangEx/commonService/generator/generate_obj_and_mock_file
// THIS FILE SHOULD NOT BE EDITED MANUALLY

package game

import (
	managerimpl "github.com/JianWangEx/commonService/generator/generate_obj_and_mock_file/fixtures/game/manager/impl"
	"github.com/JianWangEx/commonService/pkg/ioc"
)

// nolint:funlen
func init() {

	ioc.AddProvider(managerimpl.NewGameManagerImpl)

}
`
	content2 := string(outputBytes[provider2.Category][getProviderFileName(provider2)])
	assert.Equal(t, expectedForProvider2, content2)
}

func TestProviderOutputFormatter_Format_MockProvider(t *testing.T) {
	// the content for mock providers
	provider1 := ProviderInstance{
		Category:          "spn",
		PackageName:       "handler",
		PackageFullPath:   "github.com/JianWangEx/commonService/generator/generate_obj_and_mock_file/fixtures/spn/handler/mock",
		RelativeFilePath:  "",
		FileName:          "spn_handler_mock.go",
		MethodName:        "NewSpnHandlerMock",
		InterfaceFullPath: "github.com/JianWangEx/commonService/generator/generate_obj_and_mock_file/fixtures/spn/handler",
		InterfaceName:     "SpnHandler",
	}
	provider2 := ProviderInstance{
		Category:          "game",
		PackageName:       "manager",
		PackageFullPath:   "github.com/JianWangEx/commonService/generator/generate_obj_and_mock_file/fixtures/game/manager/mock",
		RelativeFilePath:  "",
		FileName:          "game_manager_mock.go",
		MethodName:        "NewGameManagerMock",
		InterfaceFullPath: "github.com/JianWangEx/commonService/generator/generate_obj_and_mock_file/fixtures/game/manager",
		InterfaceName:     "GameManager",
	}
	formatter := NewProviderOutputFormatter(configForTest(), []ProviderInstance{provider1, provider2}, "github.com/JianWangEx/commonService")
	formatter.IsMock = true
	err := formatter.Format()
	require.NoError(t, err)
	outputBytes := formatter.OutputBytes
	require.True(t, len(outputBytes) == 2, "should have two categories")
	require.True(t, len(outputBytes[provider1.Category]) > 0)
	require.True(t, len(outputBytes[provider2.Category]) > 0)
	// note the spaces and newlines, based on the template: mockInitFileTemplate
	expectedForProvider1 := `// CODE GENERATED AUTOMATICALLY BY github.com/JianWangEx/commonService/generator/generate_obj_and_mock_file
// THIS FILE SHOULD NOT BE EDITED MANUALLY

package spn

import (
	handler "github.com/JianWangEx/commonService/generator/generate_obj_and_mock_file/fixtures/spn/handler"
	handlermock "github.com/JianWangEx/commonService/generator/generate_obj_and_mock_file/fixtures/spn/handler/mock"
	"github.com/JianWangEx/commonService/pkg/ioc"
)

// nolint:funlen
func init() {

	ioc.AddMockProvider(func() handler.SpnHandler {
		return handlermock.NewSpnHandlerMockByGenerator()
	})

}
`
	content1 := string(outputBytes[provider1.Category][getProviderFileName(provider1)])
	assert.Equal(t, expectedForProvider1, content1)

	expectedForProvider2 := `// CODE GENERATED AUTOMATICALLY BY github.com/JianWangEx/commonService/generator/generate_obj_and_mock_file
// THIS FILE SHOULD NOT BE EDITED MANUALLY

package game

import (
	manager "github.com/JianWangEx/commonService/generator/generate_obj_and_mock_file/fixtures/game/manager"
	managermock "github.com/JianWangEx/commonService/generator/generate_obj_and_mock_file/fixtures/game/manager/mock"
	"github.com/JianWangEx/commonService/pkg/ioc"
)

// nolint:funlen
func init() {

	ioc.AddMockProvider(func() manager.GameManager {
		return managermock.NewGameManagerMockByGenerator()
	})

}
`
	content2 := string(outputBytes[provider2.Category][getProviderFileName(provider2)])
	assert.Equal(t, expectedForProvider2, content2)

	// based on mockConstructorFileTemplate
	mockConstructor1 := `
// CODE GENERATED AUTOMATICALLY BY github.com/JianWangEx/commonService/generator/generate_obj_and_mock_file
// THIS FILE SHOULD NOT BE EDITED MANUALLY

var spnHandler *SpnHandlerMock

func NewSpnHandlerMockByGenerator() *SpnHandlerMock {
	spnHandler = &SpnHandlerMock{}
	return spnHandler
}

func GetSpnHandlerMock() *SpnHandlerMock {
	return spnHandler
}
`
	mockConstructor2 := `
// CODE GENERATED AUTOMATICALLY BY github.com/JianWangEx/commonService/generator/generate_obj_and_mock_file
// THIS FILE SHOULD NOT BE EDITED MANUALLY

var gameManager *GameManagerMock

func NewGameManagerMockByGenerator() *GameManagerMock {
	gameManager = &GameManagerMock{}
	return gameManager
}

func GetGameManagerMock() *GameManagerMock {
	return gameManager
}
`
	mockConstructorOutputBytes := formatter.MockConstructorOutputBytes
	assert.Equal(t, mockConstructor1, string(mockConstructorOutputBytes[getMockProviderFileName(provider1, formatter.ModuleName)]))
	assert.Equal(t, mockConstructor2, string(mockConstructorOutputBytes[getMockProviderFileName(provider2, formatter.ModuleName)]))
}
