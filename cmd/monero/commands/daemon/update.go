package daemon

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/mathieumaf/go-monero/cmd/monero/display"
	"github.com/mathieumaf/go-monero/cmd/monero/options"
	"github.com/mathieumaf/go-monero/pkg/rpc/daemon"
)

type updateCommand struct {
	Command string
	Path    string
	JSON    bool
}

func (c *updateCommand) Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update",
		Short: "check for a monerod update, or download one",
		Long: `Asks the daemon whether a newer monerod release exists.

The default command is "check", which only reads that information.
"download" is state-changing: the daemon writes the archive onto its
own host, optionally under --path.`,
		RunE: c.RunE,
	}

	cmd.Flags().BoolVar(&c.JSON, "json",
		false, "whether or not to output the result as json")
	cmd.Flags().StringVar(&c.Command, "command",
		daemon.UpdateCommandCheck, `either "check" or "download"`)
	cmd.Flags().StringVar(&c.Path, "path",
		"", "directory or file path used when --command is download")

	return cmd
}

func (c *updateCommand) RunE(_ *cobra.Command, _ []string) error {
	switch c.Command {
	case daemon.UpdateCommandCheck, daemon.UpdateCommandDownload:
	default:
		return fmt.Errorf("command must be %q or %q",
			daemon.UpdateCommandCheck, daemon.UpdateCommandDownload)
	}

	ctx, cancel := options.RootOpts.Context()
	defer cancel()

	client, err := options.RootOpts.Client()
	if err != nil {
		return fmt.Errorf("client: %w", err)
	}

	params := daemon.UpdateRequestParameters{
		Command: c.Command,
		Path:    c.Path,
	}
	resp, err := client.Update(ctx, params)
	if err != nil {
		return fmt.Errorf("update: %w", err)
	}
	if c.JSON {
		return display.JSON(resp)
	}

	c.pretty(resp)
	return nil
}

func (c *updateCommand) pretty(v *daemon.UpdateResult) {
	table := display.NewTable()
	table.AddRow("Status:", v.Status)
	table.AddRow("Update available:", v.Update)
	table.AddRow("Version:", v.Version)
	table.AddRow("User URI:", v.UserURI)
	table.AddRow("Auto URI:", v.AutoURI)
	table.AddRow("Hash:", v.Hash)
	table.AddRow("Path:", v.Path)
	fmt.Println(table)
}

func init() {
	RootCommand.AddCommand((&updateCommand{}).Cmd())
}
