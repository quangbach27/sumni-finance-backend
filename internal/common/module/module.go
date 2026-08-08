package module

import (
	"context"

	"sumni-finance-backend/internal/common"
	"sumni-finance-backend/internal/common/module/contracts"
)

type Name string

type Module interface {
	Name() Name
	Init(ctx context.Context) error
	RegisterHttp(ctx context.Context, publicRouter common.EchoRouter, protectedRouter common.EchoRouter) error
	RegisterContracts(ctx context.Context, contracts *contracts.Contracts) error
}
