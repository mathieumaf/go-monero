package daemon

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/mathieumaf/go-monero/cmd/monero/display"
	"github.com/mathieumaf/go-monero/cmd/monero/options"
	"github.com/mathieumaf/go-monero/pkg/rpc/daemon"
)

type pruneBlockchainCommand struct {
	Check bool
	JSON  bool
}

func (c *pruneBlockchainCommand) Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "prune-blockchain",
		Short: "prune blockchain data (changes daemon state unless --check)",
		RunE:  c.RunE,
	}

	cmd.Flags().BoolVar(&c.Check, "check", false, "report pruning status without deleting data")
	cmd.Flags().BoolVar(&c.JSON, "json",
		false, "whether or not to output the result as json")

	return cmd
}

func (c *pruneBlockchainCommand) RunE(_ *cobra.Command, _ []string) error {
	ctx, cancel := options.RootOpts.Context()
	defer cancel()

	client, err := options.RootOpts.Client()
	if err != nil {
		return fmt.Errorf("client: %w", err)
	}

	resp, err := client.PruneBlockchain(ctx, daemon.PruneBlockchainRequestParameters{Check: c.Check})
	if err != nil {
		return fmt.Errorf("prune blockchain: %w", err)
	}

	if c.JSON {
		return display.JSON(resp)
	}

	table := display.NewTable()
	table.AddRow("Pruned:", resp.Pruned)
	table.AddRow("Pruning seed:", resp.PruningSeed)
	table.AddRow("Status:", resp.Status)
	fmt.Println(table)
	return nil
}

func init() {
	RootCommand.AddCommand((&pruneBlockchainCommand{}).Cmd())
}
