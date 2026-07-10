package wallet

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/mathieumaf/go-monero/cmd/monero/display"
	"github.com/mathieumaf/go-monero/cmd/monero/options"
	"github.com/mathieumaf/go-monero/pkg/rpc/wallet"
)

type checkReserveProofCommand struct {
	Address   string
	Message   string
	Signature string
	JSON      bool
}

func (c *checkReserveProofCommand) Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "check-reserve-proof",
		Short: "verify a reserve proof signature",
		RunE:  c.RunE,
	}

	cmd.Flags().BoolVar(&c.JSON, "json", false, "output the result as json")
	// Named "addr" to avoid clashing with the global RPC --address flag.
	cmd.Flags().StringVar(&c.Address, "addr", "",
		"public address of the proving wallet (required)")
	cmd.Flags().StringVar(&c.Message, "message", "",
		"optional message used when creating the proof")
	cmd.Flags().StringVar(&c.Signature, "signature", "",
		"reserve proof signature (required)")
	_ = cmd.MarkFlagRequired("addr")
	_ = cmd.MarkFlagRequired("signature")

	return cmd
}

func (c *checkReserveProofCommand) RunE(_ *cobra.Command, _ []string) error {
	ctx, cancel := options.RootOpts.Context()
	defer cancel()

	client, err := options.RootOpts.WalletClient()
	if err != nil {
		return fmt.Errorf("client: %w", err)
	}

	resp, err := client.CheckReserveProof(ctx, wallet.CheckReserveProofRequestParameters{
		Address:   c.Address,
		Message:   c.Message,
		Signature: c.Signature,
	})
	if err != nil {
		return fmt.Errorf("check reserve proof: %w", err)
	}

	if c.JSON {
		return display.JSON(resp)
	}

	table := display.NewTable()
	table.AddRow("Good:", resp.Good)
	table.AddRow("Total:", display.PreciseXMR(resp.Total))
	table.AddRow("Spent:", display.PreciseXMR(resp.Spent))
	fmt.Println(table)
	return nil
}

func init() {
	RootCommand.AddCommand((&checkReserveProofCommand{}).Cmd())
}
