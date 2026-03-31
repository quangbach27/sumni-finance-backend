package ledger

import (
	"fmt"
	"strings"
)

var (
	AccountingPeriodOpen  = AccoutingPeriodStatus{value: "OPEN"}
	AccountingPeriodClose = AccoutingPeriodStatus{value: "CLOSE"}
)

type AccoutingPeriodStatus struct {
	value string
}

func NewAccountingPeriodStatus(
	status string,
) (AccoutingPeriodStatus, error) {
	statusCleaned := strings.TrimSpace(strings.ToUpper(status))

	if statusCleaned == AccountingPeriodOpen.value {
		return AccountingPeriodOpen, nil
	}

	if statusCleaned == AccountingPeriodClose.value {
		return AccountingPeriodClose, nil
	}

	return AccoutingPeriodStatus{}, fmt.Errorf("unknow accounting period status: %s", status)
}

func (as AccoutingPeriodStatus) String() string { return as.value }
