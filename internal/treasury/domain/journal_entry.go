package domain

import (
	"errors"
	"fmt"
	"time"

	"sumni-finance-backend/internal/common"
	"sumni-finance-backend/internal/common/shared"

	"github.com/shopspring/decimal"
)

type JournalEntryUUID struct {
	common.UUID
}

type JournalEntryStatus struct {
	common.Enum[JournalEntryStatusValues]
}

type JournalEntryStatusValues string

func (s JournalEntryStatusValues) Values() []string {
	return []string{"RECORDED", "POSTED", "VOIDED"}
}

var (
	JournalEntryStatusRecorded = common.MustEnum[JournalEntryStatus]("RECORDED")
	JournalEntryStatusPosted   = common.MustEnum[JournalEntryStatus]("POSTED")
	JournalEntryStatusVoided   = common.MustEnum[JournalEntryStatus]("VOIDED")
)

type JournalEntry struct {
	uuid JournalEntryUUID

	amount          decimal.Decimal
	entryType       shared.EntryType
	transactionDate time.Time
	transactionNo   *string

	status JournalEntryStatus
	shared.Audit

	fundSourceUUID FundSourceUUID
	balanceBefore  decimal.Decimal
	balanceAfter   decimal.Decimal

	description *string

	voidedBy         *string
	voidedAt         *time.Time
	voidedReason     *string
	reverseEntryUUID *JournalEntryUUID
}

func NewJournalEntry(
	amount decimal.Decimal,
	entryType shared.EntryType,
	transactionDate time.Time,
	transactionNo *string,
	description *string,
	fundSourceUUID FundSourceUUID,
	balanceBefore decimal.Decimal,
) (*JournalEntry, error) {
	errDetails := []common.ErrorDetails{}

	if amount.IsZero() || amount.IsNegative() {
		errDetails = append(errDetails, common.ErrorDetails{
			EntityType: "JournalEntry",
			ErrorSlug:  "invalid-amount",
			Message:    "amount must be positive",
		})
	}
	if entryType.IsZero() {
		errDetails = append(errDetails, common.ErrorDetails{
			EntityType: "JournalEntry",
			ErrorSlug:  "empty-entry-type",
			Message:    "entry type can't be empty",
		})
	}
	if transactionDate.IsZero() {
		errDetails = append(errDetails, common.ErrorDetails{
			EntityType: "JournalEntry",
			ErrorSlug:  "empty-transaction-date",
			Message:    "transaction date can't be empty",
		})
	}
	if fundSourceUUID.IsZero() {
		errDetails = append(errDetails, common.ErrorDetails{
			EntityType: "JournalEntry",
			ErrorSlug:  "empty-fund-source-uuid",
			Message:    "fund source uuid can't be empty",
		})
	}

	if len(errDetails) != 0 {
		return nil, common.NewInvalidInputError("invalid-journal-entry", "journal entry info is invalid").WithDetails(errDetails)
	}

	return &JournalEntry{
		uuid:            JournalEntryUUID{UUID: common.NewUUIDv7()},
		amount:          amount,
		entryType:       entryType,
		transactionDate: transactionDate,
		transactionNo:   transactionNo,
		description:     description,
		status:          JournalEntryStatusRecorded,
		fundSourceUUID:  fundSourceUUID,
		balanceBefore:   balanceBefore,
		Audit:           shared.NewAudit(""),
	}, nil
}

func (j *JournalEntry) UUID() JournalEntryUUID              { return j.uuid }
func (j *JournalEntry) Amount() decimal.Decimal             { return j.amount }
func (j *JournalEntry) EntryType() shared.EntryType         { return j.entryType }
func (j *JournalEntry) TransactionDate() time.Time          { return j.transactionDate }
func (j *JournalEntry) TransactionNo() *string              { return j.transactionNo }
func (j *JournalEntry) Description() *string                { return j.description }
func (j *JournalEntry) Status() JournalEntryStatus          { return j.status }
func (j *JournalEntry) FundSourceUUID() FundSourceUUID      { return j.fundSourceUUID }
func (j *JournalEntry) VoidedBy() *string                   { return j.voidedBy }
func (j *JournalEntry) VoidedAt() *time.Time                { return j.voidedAt }
func (j *JournalEntry) VoidedReason() *string               { return j.voidedReason }
func (j *JournalEntry) BalanceBefore() decimal.Decimal      { return j.balanceBefore }
func (j *JournalEntry) BalanceAfter() decimal.Decimal       { return j.balanceAfter }
func (j *JournalEntry) ReverseEntryUUID() *JournalEntryUUID { return j.reverseEntryUUID }

func (j *JournalEntry) IsDebitSide() bool {
	return j.entryType == shared.EntryTypeDebit
}

func (j *JournalEntry) SetBalanceBefore(balance decimal.Decimal) {
	j.balanceBefore = balance
}

func (j *JournalEntry) SetBalanceAfter(balance decimal.Decimal) {
	j.balanceAfter = balance
}

func (j *JournalEntry) MarkAsPosted(updatedBy string) error {
	if j.status != JournalEntryStatusRecorded {
		return fmt.Errorf("journal entry must be in RECORDED status before posting")
	}

	j.status = JournalEntryStatusPosted
	j.SetAuditUpdate(updatedBy)

	return nil
}

func (j *JournalEntry) Void(
	voidedBy string,
	voidedAt time.Time,
	voidedReason string,
) (*JournalEntry, error) {
	if j.status == JournalEntryStatusVoided {
		return nil, errors.New("journal entry is already voided")
	}

	errDetails := []common.ErrorDetails{}
	if voidedBy == "" {
		errDetails = append(errDetails, common.ErrorDetails{
			EntityType: "JournalEntry",
			EntityID:   j.UUID().String(),
			ErrorSlug:  "empty-voided-by",
			Message:    "voided by can't be empty",
		})
	}
	if voidedAt.Before(j.CreatedAt()) {
		errDetails = append(errDetails, common.ErrorDetails{
			EntityType: "JournalEntry",
			EntityID:   j.UUID().String(),
			ErrorSlug:  "invalid-voided-at",
			Message:    "voided at can't be before created at",
		})
	}
	if voidedReason == "" {
		errDetails = append(errDetails, common.ErrorDetails{
			EntityType: "JournalEntry",
			EntityID:   j.UUID().String(),
			ErrorSlug:  "empty-voided-reason",
			Message:    "voided reason can't be empty",
		})
	}

	if len(errDetails) != 0 {
		return nil, common.NewInvalidInputError("invalid-void-info", "void info is invalid").WithDetails(errDetails)
	}

	reverse := j.generateReverseEntry(voidedBy)

	j.status = JournalEntryStatusVoided
	j.voidedAt = &voidedAt
	j.voidedBy = &voidedBy
	j.voidedReason = &voidedReason
	j.SetAuditUpdate(voidedBy)
	j.reverseEntryUUID = &reverse.uuid

	return reverse, nil
}

func (j *JournalEntry) generateReverseEntry(actorID string) *JournalEntry {
	return &JournalEntry{
		uuid:            j.UUID(),
		amount:          j.Amount(),
		entryType:       j.EntryType(),
		transactionDate: time.Now(),
		fundSourceUUID:  j.fundSourceUUID,
		Audit:           shared.NewAudit(actorID),
	}
}
