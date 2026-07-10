package wallet

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/mathieumaf/go-monero/cmd/monero/display"
	"github.com/mathieumaf/go-monero/cmd/monero/options"
	"github.com/mathieumaf/go-monero/pkg/rpc/wallet"
)

type incomingTransfersCommand struct {
	TransferType   string
	AccountIndex   uint32
	SubaddrIndices []uint
	JSON           bool
}

func (c *incomingTransfersCommand) Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "incoming-transfers",
		Short: "list owned outputs (incoming transfers / UTXOs)",
		RunE:  c.RunE,
	}

	cmd.Flags().BoolVar(&c.JSON, "json", false, "output the result as json")
	cmd.Flags().StringVar(&c.TransferType, "transfer-type", "all",
		`filter: "all", "available", or "unavailable"`)
	cmd.Flags().Uint32Var(&c.AccountIndex, "account-index", 0, "account index")
	cmd.Flags().UintSliceVar(&c.SubaddrIndices, "subaddr-index", nil,
		"restrict to subaddress indices")

	return cmd
}

func (c *incomingTransfersCommand) RunE(_ *cobra.Command, _ []string) error {
	ctx, cancel := options.RootOpts.Context()
	defer cancel()

	client, err := options.RootOpts.WalletClient()
	if err != nil {
		return fmt.Errorf("client: %w", err)
	}

	params := wallet.IncomingTransfersRequestParameters{
		TransferType: c.TransferType,
		AccountIndex: c.AccountIndex,
	}
	for _, idx := range c.SubaddrIndices {
		params.SubaddrIndices = append(params.SubaddrIndices, uint32(idx))
	}

	resp, err := client.IncomingTransfers(ctx, params)
	if err != nil {
		return fmt.Errorf("incoming transfers: %w", err)
	}

	if c.JSON {
		return display.JSON(resp)
	}

	if len(resp.Transfers) == 0 {
		fmt.Println("No incoming transfers.")
		return nil
	}

	for _, t := range resp.Transfers {
		table := display.NewTable()
		table.AddRow("Amount:", display.PreciseXMR(t.Amount))
		table.AddRow("Spent:", t.Spent)
		table.AddRow("Unlocked:", t.Unlocked)
		table.AddRow("Frozen:", t.Frozen)
		table.AddRow("Tx Hash:", t.TxHash)
		table.AddRow("Key Image:", t.KeyImage)
		table.AddRow("Pubkey:", t.Pubkey)
		table.AddRow("Global Index:", t.GlobalIndex)
		table.AddRow("Block Height:", t.BlockHeight)
		table.AddRow("Subaddr:", fmt.Sprintf("%d/%d", t.SubaddrIndex.Major, t.SubaddrIndex.Minor))
		fmt.Println(table)
		fmt.Println()
	}

	return nil
}

func init() {
	RootCommand.AddCommand((&incomingTransfersCommand{}).Cmd())
}
