/*
Copyright © 2026 Ken'ichiro Oyama <k1lowxb@gmail.com>
*/
package cmd

import (
	"encoding/json"
	"fmt"

	tailorclient "github.com/k1LoW/tailor-client-go"
	"github.com/spf13/cobra"

	"github.com/k1LoW/tack/internal/site"
)

func newSetCmd() *cobra.Command {
	var workspaceID string
	var name string
	var output string
	var allowedIPs []string

	cmd := &cobra.Command{
		Use:           "set",
		Short:         "Update settings of a Static Website (without deploying)",
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

			settings, err := site.Set(ctx, client, site.SetOptions{
				WorkspaceID:           workspaceID,
				Name:                  name,
				AllowedIPAddresses:    allowedIPs,
				AllowedIPAddressesSet: cmd.Flags().Changed("allowed-ip-addresses"),
			})
			if err != nil {
				return err
			}

			switch output {
			case "text":
				fmt.Fprintf(cmd.OutOrStdout(), "Updated %q\n", settings.Name)
				if len(settings.AllowedIPAddresses) == 0 {
					fmt.Fprintln(cmd.OutOrStdout(), "Allowed IPs: (no restriction)")
				} else {
					fmt.Fprintf(cmd.OutOrStdout(), "Allowed IPs: %v\n", settings.AllowedIPAddresses)
				}
			case "json":
				enc := json.NewEncoder(cmd.OutOrStdout())
				enc.SetIndent("", "  ")
				if err := enc.Encode(settings); err != nil {
					return fmt.Errorf("encode json: %w", err)
				}
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&workspaceID, "workspace-id", "", "Tailor Platform workspace id (uuid)")
	cmd.Flags().StringVar(&name, "name", defaultName, "static website name")
	cmd.Flags().StringVarP(&output, "output", "o", "text", "output format: text or json")
	cmd.Flags().StringSliceVar(&allowedIPs, "allowed-ip-addresses", nil, "IP/CIDR allowlist (comma-separated or repeated). Pass empty value to clear restrictions")
	cobra.CheckErr(cmd.MarkFlagRequired("workspace-id"))
	return cmd
}
