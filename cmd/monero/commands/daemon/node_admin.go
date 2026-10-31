package daemon

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/mathieumaf/go-monero/cmd/monero/display"
	"github.com/mathieumaf/go-monero/cmd/monero/options"
	"github.com/mathieumaf/go-monero/pkg/rpc/daemon"
)

// statusEndpointCommand calls a state-changing endpoint whose useful
// result is the status string.
type statusEndpointCommand struct {
	use       string
	short     string
	long      string
	errPrefix string
	stop      bool

	JSON bool
}

func (c *statusEndpointCommand) Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   c.use,
		Short: c.short,
		Long:  c.long,
		RunE:  c.RunE,
	}

	cmd.Flags().BoolVar(&c.JSON, "json",
		false, "whether or not to output the result as json")

	return cmd
}

func (c *statusEndpointCommand) RunE(_ *cobra.Command, _ []string) error {
	ctx, cancel := options.RootOpts.Context()
	defer cancel()

	client, err := options.RootOpts.Client()
	if err != nil {
		return fmt.Errorf("client: %w", err)
	}

	status, raw, err := c.call(ctx, client)
	if err != nil {
		return fmt.Errorf("%s: %w", c.errPrefix, err)
	}
	if c.JSON {
		return display.JSON(raw)
	}

	table := display.NewTable()
	table.AddRow("Status:", status)
	fmt.Println(table)
	return nil
}

func (c *statusEndpointCommand) call(
	ctx context.Context, client *daemon.Client,
) (string, interface{}, error) {
	if c.stop {
		resp, err := client.StopDaemon(ctx)
		if err != nil {
			return "", nil, err
		}
		return resp.Status, resp, nil
	}

	resp, err := client.SaveBC(ctx)
	if err != nil {
		return "", nil, err
	}
	return resp.Status, resp, nil
}

func init() {
	RootCommand.AddCommand((&statusEndpointCommand{
		use:   "save-bc",
		short: "flush the blockchain to disk",
		long: "State-changing: syncs the filesystem cache onto disk.\n\n" +
			"The chain is already saved as it changes. This call only forces that " +
			"flush so a crash is less likely to lose the latest blocks.",
		errPrefix: "save bc",
	}).Cmd())

	RootCommand.AddCommand((&statusEndpointCommand{
		use:   "stop-daemon",
		short: "ask the daemon to disconnect and shut down",
		long: "State-changing: a successful call stops the monerod process.\n\n" +
			"The daemon disconnects its peers and exits. There is no confirmation " +
			"prompt.",
		errPrefix: "stop daemon",
		stop:      true,
	}).Cmd())
}
