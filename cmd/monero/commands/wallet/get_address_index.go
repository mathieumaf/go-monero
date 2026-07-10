package wallet

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/mathieumaf/go-monero/cmd/monero/display"
	"github.com/mathieumaf/go-monero/cmd/monero/options"
	"github.com/mathieumaf/go-monero/pkg/rpc/wallet"
)

type getAddressIndexCommand struct {
	Address string
	JSON    bool
}

func (c *getAddressIndexCommand) Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get-address-index",
		Short: "resolve a (sub)address to account/address indices",
		RunE:  c.RunE,
	}

	cmd.Flags().BoolVar(&c.JSON, "json", false, "output the result as json")
	// Named "addr" to avoid clashing with the global RPC --address flag.
	cmd.Flags().StringVar(&c.Address, "addr", "", "wallet (sub)address (required)")
	_ = cmd.MarkFlagRequired("addr")

	return cmd
}

func (c *getAddressIndexCommand) RunE(_ *cobra.Command, _ []string) error {
	ctx, cancel := options.RootOpts.Context()
	defer cancel()

	client, err := options.RootOpts.WalletClient()
	if err != nil {
		return fmt.Errorf("client: %w", err)
	}

	resp, err := client.GetAddressIndex(ctx, wallet.GetAddressIndexRequestParameters{
		Address: c.Address,
	})
	if err != nil {
		return fmt.Errorf("get address index: %w", err)
	}

	if c.JSON {
		return display.JSON(resp)
	}

	table := display.NewTable()
	table.AddRow("Account Index (major):", resp.Index.Major)
	table.AddRow("Address Index (minor):", resp.Index.Minor)
	fmt.Println(table)
	return nil
}

func init() {
	RootCommand.AddCommand((&getAddressIndexCommand{}).Cmd())
}
