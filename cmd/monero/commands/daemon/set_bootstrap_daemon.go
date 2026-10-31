package daemon

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/mathieumaf/go-monero/cmd/monero/display"
	"github.com/mathieumaf/go-monero/cmd/monero/options"
	"github.com/mathieumaf/go-monero/pkg/rpc/daemon"
)

type setBootstrapDaemonCommand struct {
	Address  string
	Username string
	Password string
	Proxy    string
	JSON     bool
}

func (c *setBootstrapDaemonCommand) Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set-bootstrap-daemon",
		Short: "set the daemon used while this node is syncing",
		Long: `State-changing: later RPC calls may be proxied to the given
daemon and reported as untrusted.

An empty --address clears the bootstrap daemon. The password is sent
to the daemon and is not printed back.`,
		RunE: c.RunE,
	}

	cmd.Flags().BoolVar(&c.JSON, "json",
		false, "whether or not to output the result as json")
	cmd.Flags().StringVar(&c.Address, "address",
		"", "bootstrap daemon as host:port; empty clears it")
	cmd.Flags().StringVar(&c.Username, "rpc-username",
		"", "username for the bootstrap daemon")
	cmd.Flags().StringVar(&c.Password, "rpc-password",
		"", "password for the bootstrap daemon")
	cmd.Flags().StringVar(&c.Proxy, "proxy",
		"", "proxy used to reach the bootstrap daemon")

	return cmd
}

func (c *setBootstrapDaemonCommand) RunE(_ *cobra.Command, _ []string) error {
	ctx, cancel := options.RootOpts.Context()
	defer cancel()

	client, err := options.RootOpts.Client()
	if err != nil {
		return fmt.Errorf("client: %w", err)
	}

	params := daemon.SetBootstrapDaemonRequestParameters{
		Address:  c.Address,
		Username: c.Username,
		Password: c.Password,
		Proxy:    c.Proxy,
	}
	resp, err := client.SetBootstrapDaemon(ctx, params)
	if err != nil {
		return fmt.Errorf("set bootstrap daemon: %w", err)
	}
	if c.JSON {
		return display.JSON(resp)
	}

	table := display.NewTable()
	table.AddRow("Status:", resp.Status)
	table.AddRow("Address:", c.Address)
	fmt.Println(table)
	return nil
}

func init() {
	RootCommand.AddCommand((&setBootstrapDaemonCommand{}).Cmd())
}
