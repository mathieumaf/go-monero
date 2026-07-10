package wallet

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/mathieumaf/go-monero/cmd/monero/display"
	"github.com/mathieumaf/go-monero/cmd/monero/options"
	"github.com/mathieumaf/go-monero/pkg/rpc/wallet"
)

type queryKeyCommand struct {
	KeyType string
	JSON    bool
}

func (c *queryKeyCommand) Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "query-key",
		Short: "return a wallet key (mnemonic, view_key, or spend_key)",
		Long: `Return wallet key material.

key-type must be one of: mnemonic, view_key, spend_key.
This method is denied when monero-wallet-rpc runs in restricted mode.`,
		RunE: c.RunE,
	}

	cmd.Flags().BoolVar(&c.JSON, "json", false, "output the result as json")
	cmd.Flags().StringVar(&c.KeyType, "key-type", "",
		`key type: "mnemonic", "view_key", or "spend_key" (required)`)
	_ = cmd.MarkFlagRequired("key-type")

	return cmd
}

func (c *queryKeyCommand) RunE(_ *cobra.Command, _ []string) error {
	ctx, cancel := options.RootOpts.Context()
	defer cancel()

	client, err := options.RootOpts.WalletClient()
	if err != nil {
		return fmt.Errorf("client: %w", err)
	}

	resp, err := client.QueryKey(ctx, wallet.QueryKeyRequestParameters{
		KeyType: c.KeyType,
	})
	if err != nil {
		return fmt.Errorf("query key: %w", err)
	}

	if c.JSON {
		return display.JSON(resp)
	}

	table := display.NewTable()
	table.AddRow("Key Type:", c.KeyType)
	table.AddRow("Key:", resp.Key)
	fmt.Println(table)
	return nil
}

func init() {
	RootCommand.AddCommand((&queryKeyCommand{}).Cmd())
}
