package impl

import (
	"context"
	"github.com/JianWangEx/commonService/generator/generate_obj_and_mock_file/fixtures/spn/manager"
)

type HunterManagerImpl struct {
}

func (g HunterManagerImpl) Find(ctx context.Context, name string) (string, error) {
	return name, nil
}

// Provider()
func NewHunterMangerImpl() manager.HunterManager {
	return &HunterManagerImpl{}
}
