package wallet

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/mathieumaf/go-monero/cmd/monero/display"
	"github.com/mathieumaf/go-monero/cmd/monero/options"
	"github.com/mathieumaf/go-monero/pkg/rpc/wallet"
)

type checkTxKeyCommand struct {
	Txid    string
	TxKey   string
	Address string
	JSON    bool
}

func (c *checkTxKeyCommand) Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "check-tx-key",
		Short: "verify a transaction using its secret key",
		RunE:  c.RunE,
	}

	cmd.Flags().BoolVar(&c.JSON, "json", false, "output the result as json")
	cmd.Flags().StringVar(&c.Txid, "txid", "", "transaction id (required)")
	cmd.Flags().StringVar(&c.TxKey, "tx-key", "", "transaction secret key (required)")
	// Named "addr" to avoid clashing with the global RPC --address flag.
	cmd.Flags().StringVar(&c.Address, "addr", "",
		"destination public address (required)")
	_ = cmd.MarkFlagRequired("txid")
	_ = cmd.MarkFlagRequired("tx-key")
	_ = cmd.MarkFlagRequired("addr")

	return cmd
}

func (c *checkTxKeyCommand) RunE(_ *cobra.Command, _ []string) error {
	ctx, cancel := options.RootOpts.Context()
	defer cancel()

	client, err := options.RootOpts.WalletClient()
	if err != nil {
		return fmt.Errorf("client: %w", err)
	}

	resp, err := client.CheckTxKey(ctx, wallet.CheckTxKeyRequestParameters{
		Txid:    c.Txid,
		TxKey:   c.TxKey,
		Address: c.Address,
	})
	if err != nil {
		return fmt.Errorf("check tx key: %w", err)
	}

	if c.JSON {
		return display.JSON(resp)
	}

	table := display.NewTable()
	table.AddRow("Received:", display.PreciseXMR(resp.Received))
	table.AddRow("In Pool:", resp.InPool)
	table.AddRow("Confirmations:", resp.Confirmations)
	fmt.Println(table)
	return nil
}

func init() {
	RootCommand.AddCommand((&checkTxKeyCommand{}).Cmd())
}
