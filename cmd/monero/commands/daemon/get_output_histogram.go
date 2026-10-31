package daemon

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/mathieumaf/go-monero/cmd/monero/display"
	"github.com/mathieumaf/go-monero/cmd/monero/options"
	"github.com/mathieumaf/go-monero/pkg/rpc/daemon"
)

type getOutputHistogramCommand struct {
	Amounts      []string
	MinCount     uint64
	MaxCount     uint64
	Unlocked     bool
	RecentCutoff uint64
	JSON         bool
}

func (c *getOutputHistogramCommand) Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get-output-histogram",
		Short: "histogram of output amounts on the chain",
		RunE:  c.RunE,
	}

	cmd.Flags().StringArrayVar(&c.Amounts, "amount", nil, "output amount in atomic units")
	cmd.Flags().Uint64Var(&c.MinCount, "min-count", 0, "minimum number of instances")
	cmd.Flags().Uint64Var(&c.MaxCount, "max-count", 0, "maximum number of instances")
	cmd.Flags().BoolVar(&c.Unlocked, "unlocked", false, "only count unlocked outputs")
	cmd.Flags().Uint64Var(&c.RecentCutoff, "recent-cutoff", 0, "recent output cutoff")
	cmd.Flags().BoolVar(&c.JSON, "json",
		false, "whether or not to output the result as json")

	return cmd
}

func (c *getOutputHistogramCommand) RunE(_ *cobra.Command, _ []string) error {
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

	resp, err := client.GetOutputHistogram(ctx, daemon.GetOutputHistogramRequestParameters{
		Amounts:      amounts,
		MinCount:     c.MinCount,
		MaxCount:     c.MaxCount,
		Unlocked:     c.Unlocked,
		RecentCutoff: c.RecentCutoff,
	})
	if err != nil {
		return fmt.Errorf("get output histogram: %w", err)
	}

	if c.JSON {
		return display.JSON(resp)
	}

	c.pretty(resp)
	return nil
}

func (c *getOutputHistogramCommand) pretty(v *daemon.GetOutputHistogramResult) {
	table := display.NewTable()
	table.AddRow("AMOUNT", "TOTAL", "UNLOCKED", "RECENT")
	for _, entry := range v.Histogram {
		table.AddRow(entry.Amount, entry.TotalInstances, entry.UnlockedInstances, entry.RecentInstances)
	}
	fmt.Println(table)
}

func init() {
	RootCommand.AddCommand((&getOutputHistogramCommand{}).Cmd())
}
