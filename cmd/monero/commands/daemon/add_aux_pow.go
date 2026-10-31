package daemon

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/mathieumaf/go-monero/cmd/monero/display"
	"github.com/mathieumaf/go-monero/cmd/monero/options"
	"github.com/mathieumaf/go-monero/pkg/rpc/daemon"
)

type addAuxPowCommand struct {
	BlocktemplateBlob string
	Aux               []string
	JSON              bool
}

func (c *addAuxPowCommand) Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add-aux-pow",
		Short: "merge auxiliary proof-of-work hashes into a block template",
		RunE:  c.RunE,
	}

	cmd.Flags().StringVar(&c.BlocktemplateBlob, "blocktemplate-blob", "", "block template blob, as hex")
	cmd.Flags().StringArrayVar(&c.Aux, "aux", nil, "auxiliary pow as id:hash")
	cmd.Flags().BoolVar(&c.JSON, "json",
		false, "whether or not to output the result as json")
	_ = cmd.MarkFlagRequired("blocktemplate-blob")
	_ = cmd.MarkFlagRequired("aux")

	return cmd
}

func (c *addAuxPowCommand) RunE(_ *cobra.Command, _ []string) error {
	aux, err := parseAuxPow(c.Aux)
	if err != nil {
		return err
	}

	ctx, cancel := options.RootOpts.Context()
	defer cancel()

	client, err := options.RootOpts.Client()
	if err != nil {
		return fmt.Errorf("client: %w", err)
	}

	resp, err := client.AddAuxPow(ctx, daemon.AddAuxPowRequestParameters{
		BlocktemplateBlob: c.BlocktemplateBlob,
		AuxPow:            aux,
	})
	if err != nil {
		return fmt.Errorf("add aux pow: %w", err)
	}

	if c.JSON {
		return display.JSON(resp)
	}

	table := display.NewTable()
	table.AddRow("Merkle root:", resp.MerkleRoot)
	table.AddRow("Merkle depth:", resp.MerkleTreeDepth)
	table.AddRow("Template bytes:", len(resp.BlocktemplateBlob))
	table.AddRow("Hashing blob bytes:", len(resp.BlockhashingBlob))
	table.AddRow("Aux pow:", len(resp.AuxPow))
	table.AddRow("Status:", resp.Status)
	fmt.Println(table)
	return nil
}

func init() {
	RootCommand.AddCommand((&addAuxPowCommand{}).Cmd())
}
