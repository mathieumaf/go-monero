package daemon

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/mathieumaf/go-monero/cmd/monero/display"
	"github.com/mathieumaf/go-monero/cmd/monero/options"
	"github.com/mathieumaf/go-monero/pkg/rpc/daemon"
)

type getTxpoolBacklogCommand struct {
	JSON bool
}

func (c *getTxpoolBacklogCommand) Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get-txpool-backlog",
		Short: "transaction pool backlog entries (weight, fee, age)",
		RunE:  c.RunE,
	}

	cmd.Flags().BoolVar(&c.JSON, "json",
		false, "whether or not to output the result as json")

	return cmd
}

func (c *getTxpoolBacklogCommand) RunE(_ *cobra.Command, _ []string) error {
	ctx, cancel := options.RootOpts.Context()
	defer cancel()

	client, err := options.RootOpts.Client()
	if err != nil {
		return fmt.Errorf("client: %w", err)
	}

	resp, err := client.GetTxpoolBacklog(ctx)
	if err != nil {
		return fmt.Errorf("get txpool backlog: %w", err)
	}

	if c.JSON {
		return display.JSON(resp)
	}

	c.pretty(resp)
	return nil
}

func (c *getTxpoolBacklogCommand) pretty(v *daemon.GetTxpoolBacklogResult) {
	table := display.NewTable()
	table.AddRow("WEIGHT", "FEE", "SECONDS")
	for _, entry := range v.Entries {
		table.AddRow(entry.Weight, entry.Fee, entry.TimeInPool)
	}
	fmt.Println(table)
}

func init() {
	RootCommand.AddCommand((&getTxpoolBacklogCommand{}).Cmd())
}
