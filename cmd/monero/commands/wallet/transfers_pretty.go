package wallet

import (
	"fmt"
	"time"

	"github.com/mathieumaf/go-monero/cmd/monero/display"
	"github.com/mathieumaf/go-monero/cmd/monero/options"
	"github.com/mathieumaf/go-monero/pkg/rpc/wallet"
)

func prettyTransfer(t wallet.Transfer) {
	table := display.NewTable()
	addr := options.RootOpts.AddrFmter()(t.Address)

	table.AddRow("Txid:", t.Txid)
	table.AddRow("Type:", t.Type)
	table.AddRow("Amount:", display.PreciseXMR(t.Amount))
	table.AddRow("Fee:", display.PreciseXMR(t.Fee))
	table.AddRow("Address:", addr)
	table.AddRow("Height:", t.Height)
	table.AddRow("Confirmations:", t.Confirmations)
	table.AddRow("Locked:", t.Locked)
	table.AddRow("Unlock Time:", t.UnlockTime)
	table.AddRow("Payment ID:", t.PaymentID)
	table.AddRow("Subaddr:", fmt.Sprintf("%d/%d", t.SubaddrIndex.Major, t.SubaddrIndex.Minor))
	if t.Timestamp > 0 {
		table.AddRow("Timestamp:", time.Unix(int64(t.Timestamp), 0).UTC().Format(time.RFC3339))
	}
	if t.Note != "" {
		table.AddRow("Note:", t.Note)
	}
	if t.DoubleSpendSeen {
		table.AddRow("Double Spend Seen:", t.DoubleSpendSeen)
	}

	fmt.Println(table)

	if len(t.Destinations) > 0 {
		fmt.Println("Destinations:")
		for i, d := range t.Destinations {
			fmt.Printf("  [%d] %s  %s\n",
				i,
				options.RootOpts.AddrFmter()(d.Address),
				display.PreciseXMR(d.Amount),
			)
		}
	}

	fmt.Println()
}

func prettyPayment(p wallet.Payment) {
	table := display.NewTable()

	table.AddRow("Payment ID:", p.PaymentID)
	table.AddRow("Tx Hash:", p.TxHash)
	table.AddRow("Amount:", display.PreciseXMR(p.Amount))
	table.AddRow("Block Height:", p.BlockHeight)
	table.AddRow("Address:", options.RootOpts.AddrFmter()(p.Address))
	table.AddRow("Subaddr:", fmt.Sprintf("%d/%d", p.SubaddrIndex.Major, p.SubaddrIndex.Minor))
	table.AddRow("Unlock Time:", p.UnlockTime)
	table.AddRow("Locked:", p.Locked)

	fmt.Println(table)
	fmt.Println()
}
