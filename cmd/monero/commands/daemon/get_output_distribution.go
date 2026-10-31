package daemon

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/mathieumaf/go-monero/cmd/monero/display"
	"github.com/mathieumaf/go-monero/cmd/monero/options"
	"github.com/mathieumaf/go-monero/pkg/rpc/daemon"
)

type getOutputDistributionCommand struct {
	Amounts    []string
	FromHeight uint64
	ToHeight   uint64
	Cumulative bool
	Binary     bool
	Compress   bool
	JSON       bool
}

func (c *getOutputDistributionCommand) Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get-output-distribution",
		Short: "per-block output counts for the given amounts",
		RunE:  c.RunE,
	}

	cmd.Flags().StringArrayVar(&c.Amounts, "amount", nil, "output amount in atomic units; 0 is RingCT")
	cmd.Flags().Uint64Var(&c.FromHeight, "from-height", 0, "first height to include")
	cmd.Flags().Uint64Var(&c.ToHeight, "to-height", 0, "last height to include; 0 means the tip")
	cmd.Flags().BoolVar(&c.Cumulative, "cumulative", false, "include earlier outputs in each count")
	cmd.Flags().BoolVar(&c.Binary, "binary", false, "request monerod's packed blob instead of a JSON array")
	cmd.Flags().BoolVar(&c.Compress, "compress", false, "request compressed counts (only with --binary)")
	cmd.Flags().BoolVar(&c.JSON, "json",
		false, "whether or not to output the result as json")
	_ = cmd.MarkFlagRequired("amount")

	return cmd
}

func (c *getOutputDistributionCommand) RunE(_ *cobra.Command, _ []string) error {
	amounts, err := parseUint64List(c.Amounts)
	if err != nil {
		return err
	}

	ctx, cancel := options.RootOpts.Context()
	defer cancel()

	client, err := options.RootOpts.Client()
	if err != nil {
		return fmt.Errorf("client: %w", err)
	}

	resp, err := client.GetOutputDistribution(ctx, daemon.GetOutputDistributionRequestParameters{
		Amounts:    amounts,
		Cumulative: c.Cumulative,
		FromHeight: c.FromHeight,
		ToHeight:   c.ToHeight,
		Binary:     c.Binary,
		Compress:   c.Compress,
	})
	if err != nil {
		return fmt.Errorf("get output distribution: %w", err)
	}

	if c.JSON {
		return display.JSON(resp)
	}

	return c.pretty(resp)
}

func (c *getOutputDistributionCommand) pretty(v *daemon.GetOutputDistributionResult) error {
	table := display.NewTable()
	table.AddRow("AMOUNT", "START", "BASE", "COUNTS")
	for _, entry := range v.Distributions {
		counts, err := entry.Counts()
		if err != nil {
			return fmt.Errorf("counts: %w", err)
		}
		table.AddRow(entry.Amount, entry.StartHeight, entry.Base, len(counts))
	}
	fmt.Println(table)
	return nil
}

func init() {
	RootCommand.AddCommand((&getOutputDistributionCommand{}).Cmd())
}
