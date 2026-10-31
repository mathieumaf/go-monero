package daemon

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/mathieumaf/go-monero/cmd/monero/display"
	"github.com/mathieumaf/go-monero/cmd/monero/options"
)

type submitBlockCommand struct {
	Blobs []string
	JSON  bool
}

func (c *submitBlockCommand) Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "submit-block",
		Short: "submit a mined block (changes daemon state when accepted)",
		RunE:  c.RunE,
	}

	cmd.Flags().StringArrayVar(&c.Blobs, "blob", nil, "mined block blob, as hex")
	cmd.Flags().BoolVar(&c.JSON, "json",
		false, "whether or not to output the result as json")
	_ = cmd.MarkFlagRequired("blob")

	return cmd
}

func (c *submitBlockCommand) RunE(_ *cobra.Command, _ []string) error {
	ctx, cancel := options.RootOpts.Context()
	defer cancel()

	client, err := options.RootOpts.Client()
	if err != nil {
		return fmt.Errorf("client: %w", err)
	}

	resp, err := client.SubmitBlock(ctx, c.Blobs)
	if err != nil {
		return fmt.Errorf("submit block: %w", err)
	}

	if c.JSON {
		return display.JSON(resp)
	}

	table := display.NewTable()
	table.AddRow("Status:", resp.Status)
	table.AddRow("Block ID:", resp.BlockID)
	fmt.Println(table)
	return nil
}

func init() {
	RootCommand.AddCommand((&submitBlockCommand{}).Cmd())
}
