package wallet

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/mathieumaf/go-monero/cmd/monero/display"
	"github.com/mathieumaf/go-monero/cmd/monero/options"
	"github.com/mathieumaf/go-monero/pkg/rpc/wallet"
)

type getTransferByTxidCommand struct {
	Txid         string
	AccountIndex uint32
	JSON         bool
}

func (c *getTransferByTxidCommand) Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get-transfer-by-txid",
		Short: "show transfer details for a transaction id",
		RunE:  c.RunE,
	}

	cmd.Flags().BoolVar(&c.JSON, "json", false, "output the result as json")
	cmd.Flags().StringVar(&c.Txid, "txid", "", "transaction id (required)")
	cmd.Flags().Uint32Var(&c.AccountIndex, "account-index", 0, "account index")
	_ = cmd.MarkFlagRequired("txid")

	return cmd
}

func (c *getTransferByTxidCommand) RunE(_ *cobra.Command, _ []string) error {
	ctx, cancel := options.RootOpts.Context()
	defer cancel()

	client, err := options.RootOpts.WalletClient()
	if err != nil {
		return fmt.Errorf("client: %w", err)
	}

	resp, err := client.GetTransferByTxid(ctx, wallet.GetTransferByTxidRequestParameters{
		Txid:         c.Txid,
		AccountIndex: c.AccountIndex,
	})
	if err != nil {
		return fmt.Errorf("get transfer by txid: %w", err)
	}

	if c.JSON {
		return display.JSON(resp)
	}

	if len(resp.Transfers) > 1 {
		fmt.Printf("=== transfers (%d) ===\n\n", len(resp.Transfers))
		for _, t := range resp.Transfers {
			prettyTransfer(t)
		}
		return nil
	}

	prettyTransfer(resp.Transfer)
	return nil
}

func init() {
	RootCommand.AddCommand((&getTransferByTxidCommand{}).Cmd())
}
