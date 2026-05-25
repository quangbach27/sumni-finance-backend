package models

import (
	"context"

	"sumni-finance-backend/internal/common"
)

type OfficeRepository interface {
	OfficeByUUID(ctx context.Context, uuid OfficeUUID) (Office, error)
	SaveOffice(ctx context.Context, office Office) error
}

type OfficeUUID struct {
	common.UUID
}

type Office struct {
	UUID OfficeUUID
	Name string
}

func (o Office) Validate() error {
	errDetails := []common.ErrorDetails{}
	if o.UUID.IsZero() {
		errDetails = append(errDetails, common.ErrorDetails{
			EntityType: "office",
			ErrorSlug:  "empty-uuid",
			Message:    "office uuid can't be empty",
		})
	}

	if o.Name == "" {
		errDetails = append(errDetails, common.ErrorDetails{
			EntityType: "office",
			ErrorSlug:  "empty-name",
			Message:    "name can't be empty",
		})
	}

	if len(errDetails) != 0 {
		return common.NewInvalidInputError("validate-office", "office validation failed").WithDetails(errDetails)
	}

	return nil
}
