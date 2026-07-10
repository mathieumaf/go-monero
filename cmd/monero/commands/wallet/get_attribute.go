package wallet

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/mathieumaf/go-monero/cmd/monero/display"
	"github.com/mathieumaf/go-monero/cmd/monero/options"
	"github.com/mathieumaf/go-monero/pkg/rpc/wallet"
)

type getAttributeCommand struct {
	Key  string
	JSON bool
}

func (c *getAttributeCommand) Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get-attribute",
		Short: "get a wallet attribute value by key",
		RunE:  c.RunE,
	}

	cmd.Flags().BoolVar(&c.JSON, "json", false, "output the result as json")
	cmd.Flags().StringVar(&c.Key, "key", "", "attribute name (required)")
	_ = cmd.MarkFlagRequired("key")

	return cmd
}

func (c *getAttributeCommand) RunE(_ *cobra.Command, _ []string) error {
	ctx, cancel := options.RootOpts.Context()
	defer cancel()

	client, err := options.RootOpts.WalletClient()
	if err != nil {
		return fmt.Errorf("client: %w", err)
	}

	resp, err := client.GetAttribute(ctx, wallet.GetAttributeRequestParameters{
		Key: c.Key,
	})
	if err != nil {
		return fmt.Errorf("get attribute: %w", err)
	}

	if c.JSON {
		return display.JSON(resp)
	}

	table := display.NewTable()
	table.AddRow("Key:", c.Key)
	table.AddRow("Value:", resp.Value)
	fmt.Println(table)
	return nil
}

func init() {
	RootCommand.AddCommand((&getAttributeCommand{}).Cmd())
}
