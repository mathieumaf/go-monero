package daemon

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/mathieumaf/go-monero/cmd/monero/display"
	"github.com/mathieumaf/go-monero/cmd/monero/options"
	"github.com/mathieumaf/go-monero/pkg/rpc/daemon"
)

// hashListCommand serves the read-only hash-list endpoints. One
// implementation covers both so the two commands do not drift apart.
type hashListCommand struct {
	use       string
	short     string
	errPrefix string
	pool      bool

	JSON bool
}

func (c *hashListCommand) Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   c.use,
		Short: c.short,
		RunE:  c.RunE,
	}

	cmd.Flags().BoolVar(&c.JSON, "json",
		false, "whether or not to output the result as json")

	return cmd
}

func (c *hashListCommand) RunE(_ *cobra.Command, _ []string) error {
	ctx, cancel := options.RootOpts.Context()
	defer cancel()

	client, err := options.RootOpts.Client()
	if err != nil {
		return fmt.Errorf("client: %w", err)
	}

	status, hashes, raw, err := c.fetch(ctx, client)
	if err != nil {
		return fmt.Errorf("%s: %w", c.errPrefix, err)
	}
	if c.JSON {
		return display.JSON(raw)
	}

	table := display.NewTable()
	table.AddRow("Status:", status)
	table.AddRow("Count:", len(hashes))
	fmt.Println(table)
	for _, hash := range hashes {
		fmt.Println(hash)
	}
	return nil
}

func (c *hashListCommand) fetch(
	ctx context.Context, client *daemon.Client,
) (string, []string, interface{}, error) {
	if c.pool {
		resp, err := client.GetTransactionPoolHashes(ctx)
		if err != nil {
			return "", nil, nil, err
		}
		return resp.Status, resp.TxHashes, resp, nil
	}

	resp, err := client.GetAltBlocksHashes(ctx)
	if err != nil {
		return "", nil, nil, err
	}
	return resp.Status, resp.BlksHashes, resp, nil
}

func init() {
	RootCommand.AddCommand((&hashListCommand{
		use:       "get-alt-blocks-hashes",
		short:     "hashes of blocks that are not on the main chain",
		errPrefix: "get alt blocks hashes",
	}).Cmd())

	RootCommand.AddCommand((&hashListCommand{
		use:       "get-transaction-pool-hashes",
		short:     "transaction ids currently in the pool, without tx bodies",
		errPrefix: "get transaction pool hashes",
		pool:      true,
	}).Cmd())
}
