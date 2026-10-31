package daemon

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/mathieumaf/go-monero/cmd/monero/display"
	"github.com/mathieumaf/go-monero/cmd/monero/options"
	"github.com/mathieumaf/go-monero/pkg/rpc/daemon"
)

type setLogHashRateCommand struct {
	Visible bool
	JSON    bool
}

func (c *setLogHashRateCommand) Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set-log-hash-rate",
		Short: "show or hide hash-rate lines in the daemon log",
		Long: `State-changing: updates the daemon's logging configuration.

The daemon answers with status "NOT MINING" when it is not mining.
Pass --visible to show hash-rate lines; omit it to hide them.`,
		RunE: c.RunE,
	}

	cmd.Flags().BoolVar(&c.JSON, "json",
		false, "whether or not to output the result as json")
	cmd.Flags().BoolVar(&c.Visible, "visible",
		false, "show hash-rate lines in the daemon log")

	return cmd
}

func (c *setLogHashRateCommand) RunE(_ *cobra.Command, _ []string) error {
	ctx, cancel := options.RootOpts.Context()
	defer cancel()

	client, err := options.RootOpts.Client()
	if err != nil {
		return fmt.Errorf("client: %w", err)
	}

	params := daemon.SetLogHashRateRequestParameters{Visible: c.Visible}
	resp, err := client.SetLogHashRate(ctx, params)
	if err != nil {
		return fmt.Errorf("set log hash rate: %w", err)
	}
	if c.JSON {
		return display.JSON(resp)
	}

	table := display.NewTable()
	table.AddRow("Status:", resp.Status)
	fmt.Println(table)
	return nil
}

func init() {
	RootCommand.AddCommand((&setLogHashRateCommand{}).Cmd())
}
