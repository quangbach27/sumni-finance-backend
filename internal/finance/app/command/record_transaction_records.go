package command

import (
	"context"
	"errors"
	"sumni-finance-backend/internal/common/cqrs"
	"sumni-finance-backend/internal/common/server/httperr"
	"sumni-finance-backend/internal/finance/domain/ledger"
	"sumni-finance-backend/internal/finance/domain/wallet"

	"github.com/google/uuid"
)

type RecordTransactionRecordsCmd struct {
	WalletID           uuid.UUID
	AccountingPeridID  uuid.UUID
	YearMonth          string
	TransactionRecords []TransactionRecordCmd
}

type TransactionRecordCmd struct {
	FundProviderID  uuid.UUID
	Amount          int64
	TransactionNo   string
	TransactionType string
	Description     string
}

type RecordTransactionRecordsHandler cqrs.CommandHandler[RecordTransactionRecordsCmd]

type recordTransactionRecordsHandler struct {
	walletRepo wallet.Repository
}

func NewRecordTransactionRecordsHandler(walletRepo wallet.Repository) RecordTransactionRecordsHandler {
	return &recordTransactionRecordsHandler{walletRepo: walletRepo}
}

func (h *recordTransactionRecordsHandler) Handle(ctx context.Context, cmd RecordTransactionRecordsCmd) error {
	if len(cmd.TransactionRecords) == 0 {
		return httperr.NewIncorrectInputError(
			errors.New("empty transaction records in command"),
			"missing-transaction-records",
		)
	}
	yearMonth, err := ledger.UnmarshalYearMonthFromString(cmd.YearMonth)
	if err != nil {
		return httperr.NewIncorrectInputError(err, "failed-to-parse-yearMonth-from-string")
	}
	fpIDs, txSpecs := h.extractFpIDsAndBuildTxSpec(cmd.TransactionRecords)

	if err := h.walletRepo.CreateTransactionRecords(
		ctx,
		cmd.WalletID,
		wallet.NewProviderMatchesAnySpec(fpIDs),
		cmd.AccountingPeridID,
		func(w *wallet.Wallet) error {
			if err := w.RecordTransactions(yearMonth, txSpecs...); err != nil {
				return err
			}

			return nil
		},
	); err != nil {
		return httperr.NewUnknowError(err, "failed-to-create-ledger-records")
	}

	return nil
}

func (h *recordTransactionRecordsHandler) extractFpIDsAndBuildTxSpec(transactionRecords []TransactionRecordCmd) (
	[]uuid.UUID,
	[]wallet.TransactionSpec,
) {
	seen := make(map[uuid.UUID]struct{}, len(transactionRecords))
	fpIDs := make([]uuid.UUID, 0, len(transactionRecords))
	txSpecs := make([]wallet.TransactionSpec, 0, len(transactionRecords))

	for _, tr := range transactionRecords {
		if _, exist := seen[tr.FundProviderID]; !exist {
			seen[tr.FundProviderID] = struct{}{}
			fpIDs = append(fpIDs, tr.FundProviderID)
		}

		txSpecs = append(txSpecs, wallet.TransactionSpec{
			TransactionNo:   tr.TransactionNo,
			TransactionType: tr.TransactionType,
			Amount:          tr.Amount,
			Description:     tr.Description,
			FpID:            tr.FundProviderID,
		})
	}

	return fpIDs, txSpecs
}
