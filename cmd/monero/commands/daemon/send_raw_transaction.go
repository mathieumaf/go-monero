package daemon

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/mathieumaf/go-monero/cmd/monero/display"
	"github.com/mathieumaf/go-monero/cmd/monero/options"
	"github.com/mathieumaf/go-monero/pkg/rpc/daemon"
)

type sendRawTransactionCommand struct {
	TxHex            string
	DoNotRelay       bool
	SkipSanityChecks bool
	JSON             bool
}

func (c *sendRawTransactionCommand) Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "send-raw-transaction",
		Short: "submit a raw transaction to the daemon",
		Long: `State-changing: a transaction the daemon accepts is broadcast
to the network, unless --do-not-relay is set.

Sanity checks run by default. --skip-sanity-checks disables them.`,
		RunE: c.RunE,
	}

	cmd.Flags().BoolVar(&c.JSON, "json",
		false, "whether or not to output the result as json")
	cmd.Flags().StringVar(&c.TxHex, "tx-hex",
		"", "full transaction blob, hex-encoded")
	_ = cmd.MarkFlagRequired("tx-hex")
	cmd.Flags().BoolVar(&c.DoNotRelay, "do-not-relay",
		false, "accept the transaction without broadcasting it")
	cmd.Flags().BoolVar(&c.SkipSanityChecks, "skip-sanity-checks",
		false, "skip the daemon's tx_sanity_check")

	return cmd
}

func (c *sendRawTransactionCommand) RunE(_ *cobra.Command, _ []string) error {
	ctx, cancel := options.RootOpts.Context()
	defer cancel()

	client, err := options.RootOpts.Client()
	if err != nil {
		return fmt.Errorf("client: %w", err)
	}

	params := daemon.SendRawTransactionRequestParameters{
		TxAsHex:    c.TxHex,
		DoNotRelay: c.DoNotRelay,
	}
	if c.SkipSanityChecks {
		checks := false
		params.DoSanityChecks = &checks
	}

	resp, err := client.SendRawTransaction(ctx, params)
	if err != nil {
		return fmt.Errorf("send raw transaction: %w", err)
	}
	if c.JSON {
		return display.JSON(resp)
	}

	c.pretty(resp)
	return nil
}

func (c *sendRawTransactionCommand) pretty(v *daemon.SendRawTransactionResult) {
	table := display.NewTable()
	table.AddRow("Status:", v.Status)
	table.AddRow("Reason:", v.Reason)
	table.AddRow("Not relayed:", v.NotRelayed)
	table.AddRow("Double spend:", v.DoubleSpend)
	table.AddRow("Fee too low:", v.FeeTooLow)
	table.AddRow("Low mixin:", v.LowMixin)
	table.AddRow("Invalid input:", v.InvalidInput)
	table.AddRow("Invalid output:", v.InvalidOutput)
	table.AddRow("Too big:", v.TooBig)
	table.AddRow("Overspend:", v.Overspend)
	table.AddRow("Too few outputs:", v.TooFewOutputs)
	table.AddRow("Sanity check failed:", v.SanityCheckFailed)
	table.AddRow("Tx extra too big:", v.TxExtraTooBig)
	table.AddRow("Nonzero unlock time:", v.NonzeroUnlockTime)
	fmt.Println(table)
}

func init() {
	RootCommand.AddCommand((&sendRawTransactionCommand{}).Cmd())
}
