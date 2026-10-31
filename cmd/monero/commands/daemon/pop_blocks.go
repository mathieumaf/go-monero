package daemon

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/mathieumaf/go-monero/cmd/monero/display"
	"github.com/mathieumaf/go-monero/cmd/monero/options"
	"github.com/mathieumaf/go-monero/pkg/rpc/daemon"
)

type popBlocksCommand struct {
	Nblocks uint64
	DropTxs bool
	JSON    bool
}

func (c *popBlocksCommand) Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pop-blocks",
		Short: "remove blocks from the tip of the local chain",
		Long: `State-changing: rewrites the local chain by removing --nblocks
blocks from the tip.

By default, non-coinbase transactions from those blocks return to the
mempool. --drop-txs discards them instead. Monerod 0.18.4 and earlier
ignore that choice and keep the transactions. There is no confirmation
prompt.`,
		RunE: c.RunE,
	}

	cmd.Flags().BoolVar(&c.JSON, "json",
		false, "whether or not to output the result as json")
	cmd.Flags().Uint64Var(&c.Nblocks, "nblocks",
		0, "number of blocks to remove from the tip")
	_ = cmd.MarkFlagRequired("nblocks")
	cmd.Flags().BoolVar(&c.DropTxs, "drop-txs",
		false, "drop popped transactions instead of returning them to the pool")

	return cmd
}

func (c *popBlocksCommand) RunE(_ *cobra.Command, _ []string) error {
	ctx, cancel := options.RootOpts.Context()
	defer cancel()

	client, err := options.RootOpts.Client()
	if err != nil {
		return fmt.Errorf("client: %w", err)
	}

	params := daemon.PopBlocksRequestParameters{Nblocks: c.Nblocks}
	if c.DropTxs {
		keep := false
		params.KeepTxs = &keep
	}

	resp, err := client.PopBlocks(ctx, params)
	if err != nil {
		return fmt.Errorf("pop blocks: %w", err)
	}
	if c.JSON {
		return display.JSON(resp)
	}

	table := display.NewTable()
	table.AddRow("Status:", resp.Status)
	table.AddRow("Height:", resp.Height)
	fmt.Println(table)
	return nil
}

func init() {
	RootCommand.AddCommand((&popBlocksCommand{}).Cmd())
}
