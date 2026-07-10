package wallet

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/mathieumaf/go-monero/cmd/monero/display"
	"github.com/mathieumaf/go-monero/cmd/monero/options"
	"github.com/mathieumaf/go-monero/pkg/rpc/wallet"
)

type checkTxProofCommand struct {
	Txid      string
	Address   string
	Message   string
	Signature string
	JSON      bool
}

func (c *checkTxProofCommand) Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "check-tx-proof",
		Short: "verify a transaction proof signature",
		RunE:  c.RunE,
	}

	cmd.Flags().BoolVar(&c.JSON, "json", false, "output the result as json")
	cmd.Flags().StringVar(&c.Txid, "txid", "", "transaction id (required)")
	// Named "addr" to avoid clashing with the global RPC --address flag.
	cmd.Flags().StringVar(&c.Address, "addr", "",
		"destination public address (required)")
	cmd.Flags().StringVar(&c.Message, "message", "",
		"optional message used when creating the proof")
	cmd.Flags().StringVar(&c.Signature, "signature", "",
		"transaction proof signature (required)")
	_ = cmd.MarkFlagRequired("txid")
	_ = cmd.MarkFlagRequired("addr")
	_ = cmd.MarkFlagRequired("signature")

	return cmd
}

func (c *checkTxProofCommand) RunE(_ *cobra.Command, _ []string) error {
	ctx, cancel := options.RootOpts.Context()
	defer cancel()

	client, err := options.RootOpts.WalletClient()
	if err != nil {
		return fmt.Errorf("client: %w", err)
	}

	resp, err := client.CheckTxProof(ctx, wallet.CheckTxProofRequestParameters{
		Txid:      c.Txid,
		Address:   c.Address,
		Message:   c.Message,
		Signature: c.Signature,
	})
	if err != nil {
		return fmt.Errorf("check tx proof: %w", err)
	}

	if c.JSON {
		return display.JSON(resp)
	}

	table := display.NewTable()
	table.AddRow("Good:", resp.Good)
	table.AddRow("Received:", display.PreciseXMR(resp.Received))
	table.AddRow("In Pool:", resp.InPool)
	table.AddRow("Confirmations:", resp.Confirmations)
	fmt.Println(table)
	return nil
}

func init() {
	RootCommand.AddCommand((&checkTxProofCommand{}).Cmd())
}
