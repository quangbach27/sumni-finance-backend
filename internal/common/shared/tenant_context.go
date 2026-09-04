package shared

import "sumni-finance-backend/internal/common"

type TenantContext struct {
	tenantID string
	officeID string
}

func NewTenantContext(tenantID, officeID string) (TenantContext, error) {
	errDetails := []common.ErrorDetails{}
	if tenantID == "" {
		errDetails = append(errDetails, common.ErrorDetails{
			EntityType: "TenantContext",
			ErrorSlug:  "empty-tenant-id",
			Message:    "tenant id can't be empty",
		})
	}
	if officeID == "" {
		errDetails = append(errDetails, common.ErrorDetails{
			EntityType: "TenantContext",
			ErrorSlug:  "empty-office-id",
			Message:    "office id can't be empty",
		})
	}
	if len(errDetails) != 0 {
		return TenantContext{}, common.NewInvalidInputError("invalid-tenant-context", "tenant context is invalid").WithDetails(errDetails)
	}

	return TenantContext{tenantID: tenantID, officeID: officeID}, nil
}

func (t TenantContext) TenantID() string {
	return t.tenantID
}

func (t TenantContext) OfficeID() string {
	return t.officeID
}

func (t TenantContext) IsZero() bool {
	return t == TenantContext{}
}
