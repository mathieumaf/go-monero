package daemon

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/mathieumaf/go-monero/cmd/monero/display"
	"github.com/mathieumaf/go-monero/cmd/monero/options"
	"github.com/mathieumaf/go-monero/pkg/rpc/daemon"
)

type getMinerDataCommand struct {
	JSON bool
}

func (c *getMinerDataCommand) Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get-miner-data",
		Short: "block template fields used by miners and p2pool",
		RunE:  c.RunE,
	}

	cmd.Flags().BoolVar(&c.JSON, "json",
		false, "whether or not to output the result as json")

	return cmd
}

func (c *getMinerDataCommand) RunE(_ *cobra.Command, _ []string) error {
	ctx, cancel := options.RootOpts.Context()
	defer cancel()

	client, err := options.RootOpts.Client()
	if err != nil {
		return fmt.Errorf("client: %w", err)
	}

	resp, err := client.GetMinerData(ctx)
	if err != nil {
		return fmt.Errorf("get miner data: %w", err)
	}

	if c.JSON {
		return display.JSON(resp)
	}

	c.pretty(resp)
	return nil
}

func (c *getMinerDataCommand) pretty(v *daemon.GetMinerDataResult) {
	table := display.NewTable()
	table.AddRow("Height:", v.Height)
	table.AddRow("Major version:", v.MajorVersion)
	table.AddRow("Previous block:", v.PrevID)
	table.AddRow("Seed hash:", v.SeedHash)
	table.AddRow("Difficulty:", v.Difficulty)
	table.AddRow("Median weight:", v.MedianWeight)
	table.AddRow("Already generated coins:", v.AlreadyGeneratedCoins)
	table.AddRow("Tx backlog:", len(v.TxBacklog))
	fmt.Println(table)
}

func init() {
	RootCommand.AddCommand((&getMinerDataCommand{}).Cmd())
}
