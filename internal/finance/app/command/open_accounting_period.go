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

type OpenAccountingPeriodCmd struct {
	WalletID uuid.UUID
	Year     int
	Month    int
}

type OpenAccountingPeriodHandler cqrs.CommandHandler[OpenAccountingPeriodCmd]

type openAccountingPeriodHandler struct {
	walletRepo wallet.Repository
	ledgerRepo ledger.Repository
}

func NewOpenAccountingPeriodHandler(walletRepo wallet.Repository, ledgerRepo ledger.Repository) OpenAccountingPeriodHandler {
	return &openAccountingPeriodHandler{
		walletRepo: walletRepo,
		ledgerRepo: ledgerRepo,
	}
}

func (h *openAccountingPeriodHandler) Handle(ctx context.Context, cmd OpenAccountingPeriodCmd) error {
	w, err := h.walletRepo.GetByID(ctx, cmd.WalletID)
	if err != nil {
		return httperr.NewUnknowError(err, "failed-to-retrieve-wallet")
	}

	yearMonth, err := ledger.NewYearMonth(cmd.Month, cmd.Year)
	if err != nil {
		return httperr.NewIncorrectInputError(err, "failed-to-create-year-month")
	}

	if err = w.OpenAccountingPeriod(yearMonth); err != nil {
		return httperr.NewIncorrectInputError(err, "failed-to-open-accounting-period")
	}

	ap, exist := w.LedgerManager().FindAccountingPeriod(yearMonth)
	if !exist {
		return httperr.NewUnknowError(
			errors.New("accounting period is successful opened in domain but not found in wallet domain"),
			"faild-to-open-accounting-period",
		)
	}

	if err = h.ledgerRepo.CreateAccountingPeriod(ctx, w.ID(), ap); err != nil {
		return err
	}

	return nil
}
