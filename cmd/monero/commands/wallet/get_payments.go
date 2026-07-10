package wallet

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/mathieumaf/go-monero/cmd/monero/display"
	"github.com/mathieumaf/go-monero/cmd/monero/options"
	"github.com/mathieumaf/go-monero/pkg/rpc/wallet"
)

type getPaymentsCommand struct {
	PaymentID string
	JSON      bool
}

func (c *getPaymentsCommand) Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get-payments",
		Short: "list incoming payments for a payment id",
		RunE:  c.RunE,
	}

	cmd.Flags().BoolVar(&c.JSON, "json", false, "output the result as json")
	cmd.Flags().StringVar(&c.PaymentID, "payment-id", "",
		"16-character hex payment id (required)")
	_ = cmd.MarkFlagRequired("payment-id")

	return cmd
}

func (c *getPaymentsCommand) RunE(_ *cobra.Command, _ []string) error {
	ctx, cancel := options.RootOpts.Context()
	defer cancel()

	client, err := options.RootOpts.WalletClient()
	if err != nil {
		return fmt.Errorf("client: %w", err)
	}

	resp, err := client.GetPayments(ctx, wallet.GetPaymentsRequestParameters{
		PaymentID: c.PaymentID,
	})
	if err != nil {
		return fmt.Errorf("get payments: %w", err)
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
	RootCommand.AddCommand((&getPaymentsCommand{}).Cmd())
}
