package daemon

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/mathieumaf/go-monero/cmd/monero/display"
	"github.com/mathieumaf/go-monero/cmd/monero/options"
	"github.com/mathieumaf/go-monero/pkg/rpc/daemon"
)

type bannedCommand struct {
	Address string
	JSON    bool
}

func (c *bannedCommand) Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "banned",
		Short: "check whether an address is banned",
		RunE:  c.RunE,
	}

	cmd.Flags().StringVar(&c.Address, "address", "", "host or IP address to look up")
	cmd.Flags().BoolVar(&c.JSON, "json",
		false, "whether or not to output the result as json")
	_ = cmd.MarkFlagRequired("address")

	return cmd
}

func (c *bannedCommand) RunE(_ *cobra.Command, _ []string) error {
	ctx, cancel := options.RootOpts.Context()
	defer cancel()

	client, err := options.RootOpts.Client()
	if err != nil {
		return fmt.Errorf("client: %w", err)
	}

	resp, err := client.Banned(ctx, daemon.BannedRequestParameters{Address: c.Address})
	if err != nil {
		return fmt.Errorf("banned: %w", err)
	}

	if c.JSON {
		return display.JSON(resp)
	}

	c.pretty(resp)
	return nil
}

func (c *bannedCommand) pretty(v *daemon.BannedResult) {
	table := display.NewTable()
	table.AddRow("Banned:", v.Banned)
	table.AddRow("Seconds left:", v.Seconds)
	table.AddRow("Status:", v.Status)
	fmt.Println(table)
}

func init() {
	RootCommand.AddCommand((&bannedCommand{}).Cmd())
}
