package daemon

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/mathieumaf/go-monero/cmd/monero/display"
	"github.com/mathieumaf/go-monero/cmd/monero/options"
	"github.com/mathieumaf/go-monero/pkg/rpc/daemon"
)

type flushCacheCommand struct {
	BadBlocks bool
	JSON      bool
}

func (c *flushCacheCommand) Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "flush-cache",
		Short: "drop cached invalid blocks (changes daemon state with --bad-blocks)",
		Long: "monerod 0.18 drops the invalid-block cache only when --bad-blocks is set. " +
			"Without that flag the call leaves the cache unchanged.",
		RunE: c.RunE,
	}

	cmd.Flags().BoolVar(&c.BadBlocks, "bad-blocks", false, "drop the invalid-block cache")
	cmd.Flags().BoolVar(&c.JSON, "json",
		false, "whether or not to output the result as json")

	return cmd
}

func (c *flushCacheCommand) RunE(_ *cobra.Command, _ []string) error {
	ctx, cancel := options.RootOpts.Context()
	defer cancel()

	client, err := options.RootOpts.Client()
	if err != nil {
		return fmt.Errorf("client: %w", err)
	}

	resp, err := client.FlushCache(ctx, daemon.FlushCacheRequestParameters{BadBlocks: c.BadBlocks})
	if err != nil {
		return fmt.Errorf("flush cache: %w", err)
	}

	if c.JSON {
		return display.JSON(resp)
	}

	fmt.Println(resp.Status)
	return nil
}

func init() {
	RootCommand.AddCommand((&flushCacheCommand{}).Cmd())
}
