package wallet

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/mathieumaf/go-monero/cmd/monero/display"
	"github.com/mathieumaf/go-monero/cmd/monero/options"
	"github.com/mathieumaf/go-monero/pkg/rpc/wallet"
)

type validateAddressCommand struct {
	Address        string
	AnyNetType     bool
	AllowOpenalias bool
	JSON           bool
}

func (c *validateAddressCommand) Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "validate-address",
		Short: "validate a monero address (or OpenAlias)",
		RunE:  c.RunE,
	}

	cmd.Flags().BoolVar(&c.JSON, "json", false, "output the result as json")
	// Named "addr" to avoid clashing with the global RPC --address flag.
	cmd.Flags().StringVar(&c.Address, "addr", "", "monero address to validate (required)")
	cmd.Flags().BoolVar(&c.AnyNetType, "any-net-type", false,
		"accept addresses from any network type")
	cmd.Flags().BoolVar(&c.AllowOpenalias, "allow-openalias", false,
		"resolve OpenAlias addresses")
	_ = cmd.MarkFlagRequired("addr")

	return cmd
}

func (c *validateAddressCommand) RunE(_ *cobra.Command, _ []string) error {
	ctx, cancel := options.RootOpts.Context()
	defer cancel()

	client, err := options.RootOpts.WalletClient()
	if err != nil {
		return fmt.Errorf("client: %w", err)
	}

	resp, err := client.ValidateAddress(ctx, wallet.ValidateAddressRequestParameters{
		Address:        c.Address,
		AnyNetType:     c.AnyNetType,
		AllowOpenalias: c.AllowOpenalias,
	})
	if err != nil {
		return fmt.Errorf("validate address: %w", err)
	}

	if c.JSON {
		return display.JSON(resp)
	}

	table := display.NewTable()
	table.AddRow("Valid:", resp.Valid)
	table.AddRow("Integrated:", resp.Integrated)
	table.AddRow("Subaddress:", resp.Subaddress)
	table.AddRow("Nettype:", resp.Nettype)
	if resp.OpenaliasAddress != "" {
		table.AddRow("Openalias:", resp.OpenaliasAddress)
	}
	fmt.Println(table)
	return nil
}

func init() {
	RootCommand.AddCommand((&validateAddressCommand{}).Cmd())
}
