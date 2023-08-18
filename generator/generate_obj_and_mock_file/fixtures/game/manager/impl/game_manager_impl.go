package impl

import (
	"context"
	"github.com/JianWangEx/commonService/generator/generate_obj_and_mock_file/fixtures/game/manager"
)

type GameManagerImpl struct {
}

func (g GameManagerImpl) Find(ctx context.Context, name string) (string, error) {
	return name, nil
}

// @Provider()
func NewGameMangerImpl() manager.GameManager {
	return &GameManagerImpl{}
}
