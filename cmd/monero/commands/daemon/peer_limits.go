package daemon

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/mathieumaf/go-monero/cmd/monero/display"
	"github.com/mathieumaf/go-monero/cmd/monero/options"
	"github.com/mathieumaf/go-monero/pkg/rpc/daemon"
)

// peerLimitCommand reads or sets one peer-limit endpoint. Set defaults
// to false so invoking the command does not change the daemon.
type peerLimitCommand struct {
	use       string
	short     string
	long      string
	limitHelp string
	rowLabel  string
	errPrefix string
	incoming  bool

	Set   bool
	Limit uint32
	JSON  bool
}

func (c *peerLimitCommand) Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   c.use,
		Short: c.short,
		Long:  c.long,
		RunE:  c.RunE,
	}

	cmd.Flags().BoolVar(&c.JSON, "json",
		false, "whether or not to output the result as json")
	cmd.Flags().BoolVar(&c.Set, "set",
		false, "apply --limit instead of only reading the current limit")
	cmd.Flags().Uint32Var(&c.Limit, "limit", 0, c.limitHelp)

	return cmd
}

func (c *peerLimitCommand) RunE(cmd *cobra.Command, _ []string) error {
	if cmd.Flags().Changed("limit") && !c.Set {
		return fmt.Errorf("--limit requires --set")
	}

	ctx, cancel := options.RootOpts.Context()
	defer cancel()

	client, err := options.RootOpts.Client()
	if err != nil {
		return fmt.Errorf("client: %w", err)
	}

	status, peers, raw, err := c.call(ctx, client)
	if err != nil {
		return fmt.Errorf("%s: %w", c.errPrefix, err)
	}
	if c.JSON {
		return display.JSON(raw)
	}

	table := display.NewTable()
	table.AddRow("Status:", status)
	table.AddRow(c.rowLabel, peers)
	fmt.Println(table)
	return nil
}

func (c *peerLimitCommand) call(
	ctx context.Context, client *daemon.Client,
) (string, uint32, interface{}, error) {
	if c.incoming {
		resp, err := client.InPeers(ctx, daemon.InPeersRequestParameters{
			Set:     c.Set,
			InPeers: c.Limit,
		})
		if err != nil {
			return "", 0, nil, err
		}
		return resp.Status, resp.InPeers, resp, nil
	}

	resp, err := client.OutPeers(ctx, daemon.OutPeersRequestParameters{
		Set:      c.Set,
		OutPeers: c.Limit,
	})
	if err != nil {
		return "", 0, nil, err
	}
	return resp.Status, resp.OutPeers, resp, nil
}

func init() {
	const setNote = "The command always sends \"set\" explicitly, " +
		"because monerod treats a missing set as true and would apply the limit."

	RootCommand.AddCommand((&peerLimitCommand{
		use:   "in-peers",
		short: "read or set the maximum number of incoming peers",
		long: "Reads the daemon's maximum number of incoming public peers.\n\n" +
			"By default the current limit is only read. Pass --set with --limit to " +
			"change it. That write changes daemon state. A limit of 0 disables " +
			"incoming public peers. " + setNote,
		limitHelp: "maximum incoming public peers; applied only together with --set",
		rowLabel:  "In peers:",
		errPrefix: "in peers",
		incoming:  true,
	}).Cmd())

	RootCommand.AddCommand((&peerLimitCommand{
		use:   "out-peers",
		short: "read or set the maximum number of outgoing peers",
		long: "Reads the daemon's maximum number of outgoing public peers.\n\n" +
			"By default the current limit is only read. Pass --set with --limit to " +
			"change it. That write changes daemon state. A limit of 0 disables " +
			"outgoing public peers. " + setNote,
		limitHelp: "maximum outgoing public peers; applied only together with --set",
		rowLabel:  "Out peers:",
		errPrefix: "out peers",
	}).Cmd())
}
