package command

import (
	"context"
	"time"

	"sumni-finance-backend/internal/common"
	"sumni-finance-backend/internal/common/shared"
	"sumni-finance-backend/internal/treasury/domain"

	"github.com/shopspring/decimal"
)

type CreateTransactions struct {
	Items         []CreateTransactionItem
	TenantContext shared.TenantContext
}

type CreateTransactionItem struct {
	FundSourceUUID  domain.FundSourceUUID
	WalletUUID      *domain.WalletUUID
	EntryType       shared.EntryType
	Amount          decimal.Decimal
	Currency        shared.Currency
	Description     *string
	TransactionDate time.Time
}

func (h *Handlers) CreateTransactions(ctx context.Context, cmd CreateTransactions) error {
	if len(cmd.Items) == 0 {
		return common.NewInvalidInputError(
			"invalid-create-transactions-command",
			"create transactions command input is not valid",
		).WithDetails([]common.ErrorDetails{{
			EntityType: "CreateTransactionsCommand",
			ErrorSlug:  "empty-items",
			Message:    "items can't not be empty",
		}})
	}

	walletQueries := buildWalletQueries(cmd.Items)

	if err := h.walletRepository.CreateTransactions(
		ctx,
		cmd.TenantContext,
		walletQueries,
		func(wallets map[domain.WalletUUID]*domain.Wallet) ([]*domain.Transaction, error) {
			transactions := make([]*domain.Transaction, 0, len(cmd.Items))

			for _, item := range cmd.Items {
				amount, err := shared.NewMoney(item.Amount, item.Currency)
				if err != nil {
					return nil, common.NewInvalidInputError("invalid-amount", "%s", err.Error())
				}

				var walletUUID domain.WalletUUID
				if item.WalletUUID != nil {
					walletUUID = *item.WalletUUID
				}

				transaction, err := domain.NewTransaction(
					item.EntryType,
					amount,
					item.FundSourceUUID,
					walletUUID,
					item.Description,
					item.TransactionDate,
				)
				if err != nil {
					return nil, err
				}

				if transaction.IsRecorded() {
					wallet, exist := wallets[*item.WalletUUID]
					if !exist {
						return nil, common.NewInvalidInputError(
							"wallet-not-found",
							"wallet %s not found",
							item.WalletUUID.String(),
						)
					}

					switch item.EntryType {
					case shared.EntryTypeIn:
						err = wallet.TopUp(item.FundSourceUUID, amount)
					case shared.EntryTypeOut:
						err = wallet.Withdraw(item.FundSourceUUID, amount)
					default:
						err = common.NewInvalidInputError(
							"unsupported-entry-type",
							"%s: %s", "unsupported entry type", item.EntryType.String(),
						)
					}

					if err != nil {
						return nil, common.NewInvalidInputError("failed-to-record-transaction", "%s", err.Error())
					}
				}

				transactions = append(transactions, transaction)
			}

			return transactions, nil
		},
	); err != nil {
		return err
	}

	return nil
}

// buildWalletQueries groups items with a WalletUUID by wallet, so CreateTransactions loads only
// the wallets and fund-source allocations this batch actually needs.
func buildWalletQueries(items []CreateTransactionItem) []domain.WalletQuery {
	fundSourcesByWallet := make(map[domain.WalletUUID]map[domain.FundSourceUUID]struct{})
	walletOrder := make([]domain.WalletUUID, 0, len(items))

	for _, item := range items {
		if item.WalletUUID == nil {
			continue
		}

		if _, ok := fundSourcesByWallet[*item.WalletUUID]; !ok {
			fundSourcesByWallet[*item.WalletUUID] = make(map[domain.FundSourceUUID]struct{})
			walletOrder = append(walletOrder, *item.WalletUUID)
		}
		fundSourcesByWallet[*item.WalletUUID][item.FundSourceUUID] = struct{}{}
	}

	walletQueries := make([]domain.WalletQuery, 0, len(walletOrder))
	for _, walletUUID := range walletOrder {
		fundSourceUUIDs := make([]domain.FundSourceUUID, 0, len(fundSourcesByWallet[walletUUID]))
		for fsUUID := range fundSourcesByWallet[walletUUID] {
			fundSourceUUIDs = append(fundSourceUUIDs, fsUUID)
		}

		walletQueries = append(walletQueries, domain.WalletQuery{
			WalletUUID:      walletUUID,
			FundSourceUUIDs: fundSourceUUIDs,
		})
	}

	return walletQueries
}
