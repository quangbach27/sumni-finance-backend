package shared

import (
	"time"
)

type Audit struct {
	createdAt time.Time
	createdBy string
	updatedAt *time.Time
	updatedBy *string
}

func NewAudit(actorID string) Audit {
	if actorID == "" {
		actorID = "system"
	}

	return Audit{
		createdAt: time.Now(),
		createdBy: actorID,
	}
}

func (a *Audit) SetAuditCreate(actorID string) {
	now := time.Now()

	if actorID == "" {
		actorID = "system"
	}

	a.createdAt = now
	a.createdBy = actorID
}

func (a *Audit) SetAuditUpdate(actorID string) {
	now := time.Now()
	if actorID == "" {
		actorID = "system"
	}

	a.updatedAt = &now
	a.updatedBy = &actorID
}

func (a Audit) CreatedAt() time.Time  { return a.createdAt }
func (a Audit) CreatedBy() string     { return a.createdBy }
func (a Audit) UpdatedAt() *time.Time { return a.updatedAt }
func (a Audit) UpdatedBy() *string    { return a.updatedBy }
