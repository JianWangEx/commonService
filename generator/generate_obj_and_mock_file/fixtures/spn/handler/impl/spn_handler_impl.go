package impl

import (
	"context"
	"github.com/JianWangEx/commonService/generator/generate_obj_and_mock_file/fixtures/spn/handler"
)

type SpnHandlerImpl struct {
}

func (s SpnHandlerImpl) Deal(ctx context.Context, name string) (bool, error) {
	return true, nil
}

// @Provider()
func NewSpnHandlerImpl() handler.SpnHandler {
	return &SpnHandlerImpl{}
}
