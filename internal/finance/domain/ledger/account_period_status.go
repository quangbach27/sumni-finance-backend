package ledger

import (
	"fmt"
	"strings"
)

var (
	AccountingPeriodOpen  = AccountingPeriodStatus{value: "OPEN"}
	AccountingPeriodClose = AccountingPeriodStatus{value: "CLOSE"}
)

type AccountingPeriodStatus struct {
	value string
}

func NewAccountingPeriodStatus(
	status string,
) (AccountingPeriodStatus, error) {
	statusCleaned := strings.TrimSpace(strings.ToUpper(status))

	if statusCleaned == AccountingPeriodOpen.value {
		return AccountingPeriodOpen, nil
	}

	if statusCleaned == AccountingPeriodClose.value {
		return AccountingPeriodClose, nil
	}

	return AccountingPeriodStatus{}, fmt.Errorf("unknown accounting period status: %s", status)
}

func (as AccountingPeriodStatus) String() string { return as.value }
