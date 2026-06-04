package envelope

import (
	"context"

	"sumni-finance-backend/internal/common"
	"sumni-finance-backend/internal/common/module"
	"sumni-finance-backend/internal/common/module/contracts"
)

type Module struct{}

func NewModule() *Module {
	return &Module{}
}

func (m *Module) Name() module.Name {
	return "treasury"
}

func (m *Module) Init(ctx context.Context) error {
	return nil
}

func (m *Module) RegisterContracts(ctx context.Context, contracts *contracts.Contracts) error {
	return nil
}

func (m *Module) RegisterHttp(ctx context.Context, e common.EchoRouter) error {
	return nil
}
