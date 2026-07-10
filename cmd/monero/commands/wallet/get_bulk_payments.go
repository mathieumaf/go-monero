package wallet

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/mathieumaf/go-monero/cmd/monero/display"
	"github.com/mathieumaf/go-monero/cmd/monero/options"
	"github.com/mathieumaf/go-monero/pkg/rpc/wallet"
)

type getBulkPaymentsCommand struct {
	PaymentIDs     []string
	MinBlockHeight uint64
	JSON           bool
}

func (c *getBulkPaymentsCommand) Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get-bulk-payments",
		Short: "list incoming payments for one or more payment ids",
		RunE:  c.RunE,
	}

	cmd.Flags().BoolVar(&c.JSON, "json", false, "output the result as json")
	cmd.Flags().StringSliceVar(&c.PaymentIDs, "payment-id", nil,
		"16-character hex payment id (repeatable)")
	cmd.Flags().Uint64Var(&c.MinBlockHeight, "min-block-height", 0,
		"start looking for payments from this height")
	_ = cmd.MarkFlagRequired("payment-id")

	return cmd
}

func (c *getBulkPaymentsCommand) RunE(_ *cobra.Command, _ []string) error {
	ctx, cancel := options.RootOpts.Context()
	defer cancel()

	client, err := options.RootOpts.WalletClient()
	if err != nil {
		return fmt.Errorf("client: %w", err)
	}

	resp, err := client.GetBulkPayments(ctx, wallet.GetBulkPaymentsRequestParameters{
		PaymentIDs:     c.PaymentIDs,
		MinBlockHeight: c.MinBlockHeight,
	})
	if err != nil {
		return fmt.Errorf("get bulk payments: %w", err)
	}

	if c.JSON {
		return display.JSON(resp)
	}

	if len(resp.Payments) == 0 {
		fmt.Println("No payments found.")
		return nil
	}

	for _, p := range resp.Payments {
		prettyPayment(p)
	}
	return nil
}

func init() {
	RootCommand.AddCommand((&getBulkPaymentsCommand{}).Cmd())
}
