package ui

import (
	"fmt"
	"github.com/NodeFactoryIo/vedran/internal/payout"
	"github.com/gosuri/uitable"
)

func DisplayTransactionsStatus(transactions []*payout.TransactionDetails) {
	table := uitable.New()
	table.MaxColWidth = 80
	table.Wrap = true
	table.AddRow("To (Node)", "Amount", "Status")
	for _, tx := range transactions {
		table.AddRow(tx.To, tx.Amount.String(), tx.Status)
	}
	fmt.Println(table)
}
