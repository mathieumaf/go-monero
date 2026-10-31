package daemon

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/mathieumaf/go-monero/cmd/monero/display"
	"github.com/mathieumaf/go-monero/cmd/monero/options"
	"github.com/mathieumaf/go-monero/pkg/rpc/daemon"
)

type flushTxpoolCommand struct {
	Txids []string
	All   bool
	JSON  bool
}

func (c *flushTxpoolCommand) Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "flush-txpool",
		Short: "drop pool transactions (changes daemon state)",
		Long: "Drop specific transactions with --txid, or the entire pool with --all. " +
			"This changes daemon state.",
		RunE: c.RunE,
	}

	cmd.Flags().StringArrayVar(&c.Txids, "txid", nil, "transaction id to drop")
	cmd.Flags().BoolVar(&c.All, "all", false, "drop every transaction in the pool")
	cmd.Flags().BoolVar(&c.JSON, "json",
		false, "whether or not to output the result as json")

	return cmd
}

func (c *flushTxpoolCommand) RunE(_ *cobra.Command, _ []string) error {
	if c.All && len(c.Txids) > 0 {
		return fmt.Errorf("pass either --all or --txid")
	}
	if !c.All && len(c.Txids) == 0 {
		return fmt.Errorf("pass --all to flush the whole pool, or --txid for specific transactions")
	}

	ctx, cancel := options.RootOpts.Context()
	defer cancel()

	client, err := options.RootOpts.Client()
	if err != nil {
		return fmt.Errorf("client: %w", err)
	}

	txids := c.Txids
	if c.All {
		txids = []string{}
	}

	resp, err := client.FlushTxpool(ctx, daemon.FlushTxpoolRequestParameters{Txids: txids})
	if err != nil {
		return fmt.Errorf("flush txpool: %w", err)
	}

	if c.JSON {
		return display.JSON(resp)
	}

	fmt.Println(resp.Status)
	return nil
}

func init() {
	RootCommand.AddCommand((&flushTxpoolCommand{}).Cmd())
}
