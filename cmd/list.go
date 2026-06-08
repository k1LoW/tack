/*
Copyright © 2026 Ken'ichiro Oyama <k1lowxb@gmail.com>
*/
package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"text/tabwriter"

	tailorclient "github.com/k1LoW/tailor-client-go"
	"github.com/spf13/cobra"

	"github.com/k1LoW/tack/internal/site"
)

func newListCmd() *cobra.Command {
	var workspaceID string
	var output string

	cmd := &cobra.Command{
		Use:           "list",
		Aliases:       []string{"ls"},
		Short:         "List Static Websites in a workspace",
		Args:          cobra.NoArgs,
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			switch output {
			case "text", "json":
			default:
				return fmt.Errorf("invalid --output %q (must be text or json)", output)
			}

			ctx := cmd.Context()
			client, err := tailorclient.New(ctx)
			if err != nil {
				return fmt.Errorf("init client: %w", err)
			}

			items, err := site.List(ctx, client, site.ListOptions{WorkspaceID: workspaceID})
			if err != nil {
				return err
			}

			switch output {
			case "text":
				return renderListText(cmd.OutOrStdout(), items)
			case "json":
				enc := json.NewEncoder(cmd.OutOrStdout())
				enc.SetIndent("", "  ")
				if len(items) == 0 {
					items = []*site.Settings{}
				}
				if err := enc.Encode(items); err != nil {
					return fmt.Errorf("encode json: %w", err)
				}
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&workspaceID, "workspace-id", "", "Tailor Platform workspace id (uuid)")
	cmd.Flags().StringVarP(&output, "output", "o", "text", "output format: text or json")
	cobra.CheckErr(cmd.MarkFlagRequired("workspace-id"))
	return cmd
}

func renderListText(w io.Writer, items []*site.Settings) error {
	if len(items) == 0 {
		fmt.Fprintln(w, "(no static websites)")
		return nil
	}
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "NAME\tURL\tALLOWED IPS")
	for _, it := range items {
		ips := "-"
		if len(it.AllowedIPAddresses) > 0 {
			ips = strings.Join(it.AllowedIPAddresses, ",")
		}
		fmt.Fprintf(tw, "%s\t%s\t%s\n", it.Name, it.URL, ips)
	}
	return tw.Flush()
}
