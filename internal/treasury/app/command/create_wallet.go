package command

import "context"

type CreateWallet struct {
}

func (h *Handlers) CreateWallet(ctx context.Context, cmd CreateWallet) error {
	return nil
}
