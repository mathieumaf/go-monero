package wallet

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/mathieumaf/go-monero/cmd/monero/display"
	"github.com/mathieumaf/go-monero/cmd/monero/options"
	"github.com/mathieumaf/go-monero/pkg/rpc/wallet"
)

type getTransfersCommand struct {
	In             bool
	Out            bool
	Pending        bool
	Failed         bool
	Pool           bool
	FilterByHeight bool
	MinHeight      uint64
	MaxHeight      uint64
	AccountIndex   uint32
	SubaddrIndices []uint
	AllAccounts    bool

	JSON bool
}

func (c *getTransfersCommand) Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get-transfers",
		Short: "list wallet transfers (in/out/pending/failed/pool)",
		RunE:  c.RunE,
	}

	cmd.Flags().BoolVar(&c.JSON, "json", false, "output the result as json")
	cmd.Flags().BoolVar(&c.In, "in", true, "include incoming transfers")
	cmd.Flags().BoolVar(&c.Out, "out", true, "include outgoing transfers")
	cmd.Flags().BoolVar(&c.Pending, "pending", true, "include pending transfers")
	cmd.Flags().BoolVar(&c.Failed, "failed", false, "include failed transfers")
	cmd.Flags().BoolVar(&c.Pool, "pool", true, "include pool transfers")
	cmd.Flags().BoolVar(&c.FilterByHeight, "filter-by-height", false,
		"restrict results to --min-height / --max-height")
	cmd.Flags().Uint64Var(&c.MinHeight, "min-height", 0, "minimum block height")
	cmd.Flags().Uint64Var(&c.MaxHeight, "max-height", 0, "maximum block height")
	cmd.Flags().Uint32Var(&c.AccountIndex, "account-index", 0, "account index")
	cmd.Flags().UintSliceVar(&c.SubaddrIndices, "subaddr-index", nil,
		"restrict to subaddress indices")
	cmd.Flags().BoolVar(&c.AllAccounts, "all-accounts", false, "query all accounts")

	return cmd
}

func (c *getTransfersCommand) RunE(_ *cobra.Command, _ []string) error {
	ctx, cancel := options.RootOpts.Context()
	defer cancel()

	client, err := options.RootOpts.WalletClient()
	if err != nil {
		return fmt.Errorf("client: %w", err)
	}

	params := wallet.GetTransfersRequestParameters{
		In:             c.In,
		Out:            c.Out,
		Pending:        c.Pending,
		Failed:         c.Failed,
		Pool:           c.Pool,
		FilterByHeight: c.FilterByHeight,
		MinHeight:      c.MinHeight,
		MaxHeight:      c.MaxHeight,
		AccountIndex:   c.AccountIndex,
		AllAccounts:    c.AllAccounts,
	}
	for _, idx := range c.SubaddrIndices {
		params.SubaddrIndices = append(params.SubaddrIndices, uint32(idx))
	}

	resp, err := client.GetTransfers(ctx, params)
	if err != nil {
		return fmt.Errorf("get transfers: %w", err)
	}

	if c.JSON {
		return display.JSON(resp)
	}

	c.pretty(resp)
	return nil
}

func (c *getTransfersCommand) pretty(v *wallet.GetTransfersResult) {
	printGroup := func(name string, transfers []wallet.Transfer) {
		if len(transfers) == 0 {
			return
		}
		fmt.Printf("=== %s (%d) ===\n\n", name, len(transfers))
		for _, t := range transfers {
			prettyTransfer(t)
		}
	}

	printGroup("in", v.In)
	printGroup("out", v.Out)
	printGroup("pending", v.Pending)
	printGroup("failed", v.Failed)
	printGroup("pool", v.Pool)

	total := len(v.In) + len(v.Out) + len(v.Pending) + len(v.Failed) + len(v.Pool)
	if total == 0 {
		fmt.Println("No transfers matched the selected filters.")
	}
}

func init() {
	RootCommand.AddCommand((&getTransfersCommand{}).Cmd())
}
