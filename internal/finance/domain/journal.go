package domain

import (
	"errors"
	"time"

	"sumni-finance-backend/internal/common"
	"sumni-finance-backend/internal/common/shared"

	"github.com/shopspring/decimal"
)

var (
	JournalEntryStatusDraft  = common.MustEnum[JournalEntryStatus]("DRAFT")
	JournalEntryStatusVoided = common.MustEnum[JournalEntryStatus]("VOIDED")
	JournalEntryStatusPosted = common.MustEnum[JournalEntryStatus]("POSTED")
)

type JournalEntryStatus struct {
	common.Enum[JournalEntryStatusValues]
}

type JournalEntryStatusValues string

func (JournalEntryStatusValues) Values() []string {
	return []string{"DRAFT", "VOIDED", "POSTED"}
}

type JournalEntryUUID struct {
	common.UUID
}

type JournalEntry struct {
	uuid             JournalEntryUUID
	fundProviderUUID FundProviderUUID

	amount          decimal.Decimal
	entryType       shared.EntryType
	transactionDate time.Time
	transactionNo   *string

	status JournalEntryStatus

	reverseEntry *JournalEntryUUID
	voidReason   *string
	voidDate     *time.Time
	voidBy       *string
}

func NewJournalEntry(
	fundproviderUUID FundProviderUUID,
	amount decimal.Decimal,
	entryType shared.EntryType,
	transactionDate time.Time,
	transactionNo *string,
) (*JournalEntry, error) {
	errDetails := []common.ErrorDetails{}

	if fundproviderUUID.IsZero() {
		errDetails = append(errDetails, common.ErrorDetails{
			EntityType: "JournalEntry",
			ErrorSlug:  "",
			Message:    "",
		})
	}
	if amount.IsZero() {
		errDetails = append(errDetails, common.ErrorDetails{
			EntityType: "JournalEntry",
			ErrorSlug:  "",
			Message:    "",
		})
	}
	if entryType.IsZero() {
		errDetails = append(errDetails, common.ErrorDetails{
			EntityType: "JournalEntry",
			ErrorSlug:  "",
			Message:    "",
		})
	}
	if transactionDate.IsZero() {
		errDetails = append(errDetails, common.ErrorDetails{
			EntityType: "JournalEntry",
			ErrorSlug:  "",
			Message:    "",
		})
	}

	if len(errDetails) != 0 {
		return nil, common.NewInvalidInputError("invalid-journal-entry", "journal entry is not valid").WithDetails(errDetails)
	}

	return &JournalEntry{
		uuid:             JournalEntryUUID{UUID: common.NewUUIDv7()},
		fundProviderUUID: fundproviderUUID,
		amount:           amount,
		entryType:        entryType,
		transactionDate:  transactionDate,
		transactionNo:    transactionNo,
		status:           JournalEntryStatusDraft,
	}, nil
}

func NewReverseJournalEntry(prevEntry *JournalEntry) (*JournalEntry, error) {
	if prevEntry == nil {
		return nil, errors.New("previous journal entry can't be empty")
	}

	if prevEntry.IsDraft() {
		return nil, errors.New("previous entry should be posted before create reverse journal entry")
	}

	return &JournalEntry{
		uuid:             JournalEntryUUID{UUID: common.NewUUIDv7()},
		fundProviderUUID: prevEntry.fundProviderUUID,
		amount:           prevEntry.amount,
		entryType:        prevEntry.entryType.Reverse(),
		transactionDate:  time.Now().UTC(),
		status:           JournalEntryStatusDraft,
	}, nil
}

func (je *JournalEntry) UUID() JournalEntryUUID             { return je.uuid }
func (je *JournalEntry) FundProviderUUID() FundProviderUUID { return je.fundProviderUUID }
func (je *JournalEntry) Amount() decimal.Decimal            { return je.amount }
func (je *JournalEntry) EntryType() shared.EntryType        { return je.entryType }
func (je *JournalEntry) TransactionDate() time.Time         { return je.transactionDate }
func (je *JournalEntry) TransactionNo() *string             { return je.transactionNo }
func (je *JournalEntry) Status() JournalEntryStatus         { return je.status }
func (je *JournalEntry) ReverseEntry() *JournalEntryUUID    { return je.reverseEntry }
func (je *JournalEntry) VoidReason() *string                { return je.voidReason }
func (je *JournalEntry) VoidDate() *time.Time               { return je.voidDate }
func (je *JournalEntry) VoidBy() *string                    { return je.voidBy }

func (je *JournalEntry) IsVoided() bool {
	return je.status.Equal(JournalEntryStatusVoided.Enum)
}

func (je *JournalEntry) IsDraft() bool {
	return je.status.Equal(JournalEntryStatusDraft.Enum)
}

func (je *JournalEntry) IsPosted() bool {
	return je.status.Equal(JournalEntryStatusPosted.Enum)
}

func (je *JournalEntry) Post() error {
	if !je.IsDraft() {
		return errors.New("journal entry has already posted")
	}

	je.status = JournalEntryStatusPosted
	return nil
}

func (je *JournalEntry) Void(
	reverseEntry JournalEntryUUID,
	voidReason string,
	voidBy string,
) error {
	if reverseEntry.IsZero() {
		return errors.New("reverse entry can't be empty")
	}
	if voidReason == "" {
		return errors.New("void reason can't be empty")
	}
	if voidBy == "" {
		return errors.New("void by can't be empty")
	}
	if je.IsVoided() {
		return errors.New("journal entry has already voided")
	}
	if je.IsDraft() {
		return errors.New("draft entry can't be voided. You can update the draft entry directly")
	}

	je.status = JournalEntryStatusVoided
	je.reverseEntry = common.ToPtr(reverseEntry)
	je.voidReason = common.ToPtr(voidReason)
	je.voidBy = common.ToPtr(voidBy)
	je.voidDate = common.ToPtr(time.Now().UTC())

	return nil
}
