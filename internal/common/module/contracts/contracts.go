package contracts

import (
	"errors"

	identityModule "sumni-finance-backend/internal/identity/api/module/client"
)

type Contracts struct {
	identityModule.Identity
}

func (c *Contracts) Verify() error {
	var err error
	if c.Identity == nil {
		err = errors.Join(err, errors.New("identtity module contract is empty"))
	}

	return err
}
