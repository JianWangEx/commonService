package generate_obj_and_mock_file

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestPreprocessor_Process(t *testing.T) {
	p := NewPreprocessor(configForTest())
	p.modFilePathForTest = "../../go.mod"
	pkgPathToName, pkgNameToFullPath, moduleName, err := p.Process()
	require.NoError(t, err)
	expectedPkgPathToName := map[string]string{
		"./generator/generate_obj_and_mock_file/fixtures/game/manager":      "gamemanager",
		"./generator/generate_obj_and_mock_file/fixtures/game/manager/impl": "managerimpl",
		"./generator/generate_obj_and_mock_file/fixtures/spn/handler":       "spnhandler",
		"./generator/generate_obj_and_mock_file/fixtures/spn/handler/impl":  "handlerimpl",
		"./generator/generate_obj_and_mock_file/fixtures/spn/manager":       "spnmanager",
		"./generator/generate_obj_and_mock_file/fixtures/spn/manager/impl":  "spnmanagerimpl", // it conflicts with game/manager/impl, so it has prefix 'spn'
	}
	expectedPkgNameToFullPath := map[string]string{
		"gamemanager":    "github.com/JianWangEx/commonService/generator/generate_obj_and_mock_file/fixtures/game/manager",
		"managerimpl":    "github.com/JianWangEx/commonService/generator/generate_obj_and_mock_file/fixtures/game/manager/impl",
		"spnhandler":     "github.com/JianWangEx/commonService/generator/generate_obj_and_mock_file/fixtures/spn/handler",
		"handlerimpl":    "github.com/JianWangEx/commonService/generator/generate_obj_and_mock_file/fixtures/spn/handler/impl",
		"spnmanager":     "github.com/JianWangEx/commonService/generator/generate_obj_and_mock_file/fixtures/spn/manager",
		"spnmanagerimpl": "github.com/JianWangEx/commonService/generator/generate_obj_and_mock_file/fixtures/spn/manager/impl",
	}
	assert.Equal(t, "github.com/JianWangEx/commonService", moduleName)
	assert.EqualValues(t, expectedPkgPathToName, pkgPathToName)
	assert.EqualValues(t, expectedPkgNameToFullPath, pkgNameToFullPath)
}

func TestPreprocessor_Process_With_Filter(t *testing.T) {
	p := NewPreprocessor(configForTest())
	p.modFilePathForTest = "../../go.mod"
	p.config.FilterKeyword = "/spn/" // only need to process spn
	pkgPathToName, pkgNameToFullPath, moduleName, err := p.Process()
	require.NoError(t, err)
	expectedPkgPathToName := map[string]string{
		"./generator/generate_obj_and_mock_file/fixtures/spn/handler":      "spnhandler",
		"./generator/generate_obj_and_mock_file/fixtures/spn/handler/impl": "handlerimpl",
		"./generator/generate_obj_and_mock_file/fixtures/spn/manager":      "spnmanager",
		"./generator/generate_obj_and_mock_file/fixtures/spn/manager/impl": "spnmanagerimpl", // it conflicts with game/manager/impl, so it has prefix 'spn'
	}
	expectedPkgNameToFullPath := map[string]string{
		"spnhandler":     "github.com/JianWangEx/commonService/generator/generate_obj_and_mock_file/fixtures/spn/handler",
		"handlerimpl":    "github.com/JianWangEx/commonService/generator/generate_obj_and_mock_file/fixtures/spn/handler/impl",
		"spnmanager":     "github.com/JianWangEx/commonService/generator/generate_obj_and_mock_file/fixtures/spn/manager",
		"spnmanagerimpl": "github.com/JianWangEx/commonService/generator/generate_obj_and_mock_file/fixtures/spn/manager/impl",
	}
	assert.Equal(t, "github.com/JianWangEx/commonService", moduleName)
	assert.EqualValues(t, expectedPkgPathToName, pkgPathToName)
	assert.EqualValues(t, expectedPkgNameToFullPath, pkgNameToFullPath)
}

func printMap[K comparable, V any](t *testing.T, title string, dict map[K]V) {
	t.Log("--------------", title, "--------------")
	for k, v := range dict {
		fmt.Println(fmt.Sprintf(`"%v": "%v",`, k, v))
	}
}

func configForTest() Config {
	config := DefaultConfig
	config.ScanPkgs = []string{
		"/generator/generate_obj_and_mock_file/fixtures/game/",
		"/generator/generate_obj_and_mock_file/fixtures/spn/",
	}
	return config
}
