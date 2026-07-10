package wallet

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/mathieumaf/go-monero/cmd/monero/display"
	"github.com/mathieumaf/go-monero/cmd/monero/options"
)

type getLanguagesCommand struct {
	JSON bool
}

func (c *getLanguagesCommand) Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get-languages",
		Short: "list available wallet seed languages",
		RunE:  c.RunE,
	}

	cmd.Flags().BoolVar(&c.JSON, "json", false, "output the result as json")

	return cmd
}

func (c *getLanguagesCommand) RunE(_ *cobra.Command, _ []string) error {
	ctx, cancel := options.RootOpts.Context()
	defer cancel()

	client, err := options.RootOpts.WalletClient()
	if err != nil {
		return fmt.Errorf("client: %w", err)
	}

	resp, err := client.GetLanguages(ctx)
	if err != nil {
		return fmt.Errorf("get languages: %w", err)
	}

	if c.JSON {
		return display.JSON(resp)
	}

	table := display.NewTable()
	table.AddRow("LANGUAGE", "LOCAL NAME")
	for i, lang := range resp.Languages {
		local := ""
		if i < len(resp.LanguagesLocal) {
			local = resp.LanguagesLocal[i]
		}
		table.AddRow(lang, local)
	}
	fmt.Println(table)
	return nil
}

func init() {
	RootCommand.AddCommand((&getLanguagesCommand{}).Cmd())
}
