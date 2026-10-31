package daemon

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/mathieumaf/go-monero/cmd/monero/display"
	"github.com/mathieumaf/go-monero/cmd/monero/options"
	"github.com/mathieumaf/go-monero/pkg/rpc/daemon"
)

type isKeyImageSpentCommand struct {
	KeyImages []string
	JSON      bool
}

func (c *isKeyImageSpentCommand) Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "is-key-image-spent",
		Short: "check whether key images are spent",
		Long: `Reports, for each key image, whether it is unspent (0), spent on
the main chain (1), or spent by a transaction still in the pool (2).`,
		RunE: c.RunE,
	}

	cmd.Flags().BoolVar(&c.JSON, "json",
		false, "whether or not to output the result as json")
	cmd.Flags().StringArrayVar(&c.KeyImages, "key-image",
		nil, "hex-encoded key image to check; repeat the flag to check several")
	_ = cmd.MarkFlagRequired("key-image")

	return cmd
}

func (c *isKeyImageSpentCommand) RunE(_ *cobra.Command, _ []string) error {
	ctx, cancel := options.RootOpts.Context()
	defer cancel()

	client, err := options.RootOpts.Client()
	if err != nil {
		return fmt.Errorf("client: %w", err)
	}

	params := daemon.IsKeyImageSpentRequestParameters{KeyImages: c.KeyImages}
	resp, err := client.IsKeyImageSpent(ctx, params)
	if err != nil {
		return fmt.Errorf("is key image spent: %w", err)
	}
	if c.JSON {
		return display.JSON(resp)
	}

	c.pretty(resp)
	return nil
}

func (c *isKeyImageSpentCommand) pretty(v *daemon.IsKeyImageSpentResult) {
	table := display.NewTable()
	table.AddRow("KEY IMAGE", "STATUS")
	for i, image := range c.KeyImages {
		status := "unknown"
		if i < len(v.SpentStatus) {
			status = keyImageStatusText(v.SpentStatus[i])
		}
		table.AddRow(image, status)
	}
	fmt.Println(table)
}

func keyImageStatusText(status daemon.KeyImageSpentStatus) string {
	switch status {
	case daemon.KeyImageUnspent:
		return "unspent"
	case daemon.KeyImageSpentInBlockchain:
		return "spent in blockchain"
	case daemon.KeyImageSpentInPool:
		return "spent in pool"
	default:
		return fmt.Sprintf("unknown (%d)", status)
	}
}

func init() {
	RootCommand.AddCommand((&isKeyImageSpentCommand{}).Cmd())
}
