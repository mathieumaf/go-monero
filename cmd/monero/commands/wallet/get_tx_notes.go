package wallet

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/mathieumaf/go-monero/cmd/monero/display"
	"github.com/mathieumaf/go-monero/cmd/monero/options"
	"github.com/mathieumaf/go-monero/pkg/rpc/wallet"
)

type getTxNotesCommand struct {
	Txids []string
	JSON  bool
}

func (c *getTxNotesCommand) Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get-tx-notes",
		Short: "get notes attached to transaction ids",
		RunE:  c.RunE,
	}

	cmd.Flags().BoolVar(&c.JSON, "json", false, "output the result as json")
	cmd.Flags().StringSliceVar(&c.Txids, "txid", nil, "transaction id (repeatable)")
	_ = cmd.MarkFlagRequired("txid")

	return cmd
}

func (c *getTxNotesCommand) RunE(_ *cobra.Command, _ []string) error {
	ctx, cancel := options.RootOpts.Context()
	defer cancel()

	client, err := options.RootOpts.WalletClient()
	if err != nil {
		return fmt.Errorf("client: %w", err)
	}

	resp, err := client.GetTxNotes(ctx, wallet.GetTxNotesRequestParameters{
		Txids: c.Txids,
	})
	if err != nil {
		return fmt.Errorf("get tx notes: %w", err)
	}

	if c.JSON {
		return display.JSON(resp)
	}

	table := display.NewTable()
	for i, txid := range c.Txids {
		note := ""
		if i < len(resp.Notes) {
			note = resp.Notes[i]
		}
		if note == "" {
			note = "(none)"
		}
		table.AddRow(txid, note)
	}
	fmt.Println(table)
	return nil
}

func init() {
	RootCommand.AddCommand((&getTxNotesCommand{}).Cmd())
}
