package wallet

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/mathieumaf/go-monero/cmd/monero/display"
	"github.com/mathieumaf/go-monero/cmd/monero/options"
	"github.com/mathieumaf/go-monero/pkg/rpc/wallet"
)

type checkSpendProofCommand struct {
	Txid      string
	Message   string
	Signature string
	JSON      bool
}

func (c *checkSpendProofCommand) Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "check-spend-proof",
		Short: "verify a spend proof signature",
		RunE:  c.RunE,
	}

	cmd.Flags().BoolVar(&c.JSON, "json", false, "output the result as json")
	cmd.Flags().StringVar(&c.Txid, "txid", "", "transaction id (required)")
	cmd.Flags().StringVar(&c.Message, "message", "",
		"optional message used when creating the proof")
	cmd.Flags().StringVar(&c.Signature, "signature", "",
		"spend proof signature (required)")
	_ = cmd.MarkFlagRequired("txid")
	_ = cmd.MarkFlagRequired("signature")

	return cmd
}

func (c *checkSpendProofCommand) RunE(_ *cobra.Command, _ []string) error {
	ctx, cancel := options.RootOpts.Context()
	defer cancel()

	client, err := options.RootOpts.WalletClient()
	if err != nil {
		return fmt.Errorf("client: %w", err)
	}

	resp, err := client.CheckSpendProof(ctx, wallet.CheckSpendProofRequestParameters{
		Txid:      c.Txid,
		Message:   c.Message,
		Signature: c.Signature,
	})
	if err != nil {
		return fmt.Errorf("check spend proof: %w", err)
	}

	if c.JSON {
		return display.JSON(resp)
	}

	table := display.NewTable()
	table.AddRow("Good:", resp.Good)
	fmt.Println(table)
	return nil
}

func init() {
	RootCommand.AddCommand((&checkSpendProofCommand{}).Cmd())
}
