package daemon

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/mathieumaf/go-monero/cmd/monero/display"
	"github.com/mathieumaf/go-monero/cmd/monero/options"
	"github.com/mathieumaf/go-monero/pkg/rpc/daemon"
)

type calcPowCommand struct {
	MajorVersion uint8
	Height       uint64
	BlockBlob    string
	SeedHash     string
	JSON         bool
}

func (c *calcPowCommand) Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "calc-pow",
		Short: "calculate the proof-of-work hash of a block candidate",
		RunE:  c.RunE,
	}

	cmd.Flags().Uint8Var(&c.MajorVersion, "major-version", 0, "block major version")
	cmd.Flags().Uint64Var(&c.Height, "height", 0, "block height")
	cmd.Flags().StringVar(&c.BlockBlob, "block-blob", "", "block blob, as hex")
	cmd.Flags().StringVar(&c.SeedHash, "seed-hash", "", "RandomX seed hash")
	cmd.Flags().BoolVar(&c.JSON, "json",
		false, "whether or not to output the result as json")
	_ = cmd.MarkFlagRequired("block-blob")
	_ = cmd.MarkFlagRequired("seed-hash")

	return cmd
}

func (c *calcPowCommand) RunE(_ *cobra.Command, _ []string) error {
	ctx, cancel := options.RootOpts.Context()
	defer cancel()

	client, err := options.RootOpts.Client()
	if err != nil {
		return fmt.Errorf("client: %w", err)
	}

	resp, err := client.CalcPow(ctx, daemon.CalcPowRequestParameters{
		MajorVersion: c.MajorVersion,
		Height:       c.Height,
		BlockBlob:    c.BlockBlob,
		SeedHash:     c.SeedHash,
	})
	if err != nil {
		return fmt.Errorf("calc pow: %w", err)
	}

	if c.JSON {
		return display.JSON(resp)
	}

	fmt.Println(resp)
	return nil
}

func init() {
	RootCommand.AddCommand((&calcPowCommand{}).Cmd())
}
