package models

import "sumni-finance-backend/internal/common"

type JournalEntryUUID struct {
	common.UUID
}

type JournalEntry struct {
	UUID JournalEntryUUID
}
