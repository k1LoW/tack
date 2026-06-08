/*
Copyright © 2026 Ken'ichiro Oyama <k1lowxb@gmail.com>
*/
package cmd

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/spf13/cobra"

	"github.com/k1LoW/tack/internal/site"
)

const defaultName = "tack"

func newUpCmd() *cobra.Command {
	var workspaceID string
	var name string
	var output string
	var allowedIPs []string

	cmd := &cobra.Command{
		Use:           "up <path>",
		Short:         "Deploy a directory or file to Tailor Platform static website hosting",
		Args:          cobra.ExactArgs(1),
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			switch output {
			case "text", "json":
			default:
				return fmt.Errorf("invalid --output %q (must be text or json)", output)
			}

			ctx := cmd.Context()
			client, err := newTailorClient(ctx)
			if err != nil {
				return fmt.Errorf("init client: %w", err)
			}

			var progress io.Writer
			if output == "text" {
				progress = cmd.OutOrStdout()
			}
			result, err := site.Deploy(ctx, client, site.DeployOptions{
				WorkspaceID:           workspaceID,
				Name:                  name,
				Path:                  args[0],
				AllowedIPAddresses:    allowedIPs,
				AllowedIPAddressesSet: cmd.Flags().Changed("allowed-ip-addresses"),
				Progress:              progress,
			})
			if err != nil {
				return err
			}

			switch output {
			case "text":
				fmt.Fprintf(cmd.OutOrStdout(), "Deployed %q\nURL: %s\n", result.Name, result.URL)
			case "json":
				enc := json.NewEncoder(cmd.OutOrStdout())
				enc.SetIndent("", "  ")
				if err := enc.Encode(result); err != nil {
					return fmt.Errorf("encode json: %w", err)
				}
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&workspaceID, "workspace-id", "", "Tailor Platform workspace id (uuid)")
	cmd.Flags().StringVar(&name, "name", defaultName, "static website name")
	cmd.Flags().StringVarP(&output, "output", "o", "text", "output format: text or json")
	cmd.Flags().StringSliceVar(&allowedIPs, "allowed-ip-addresses", nil, "IP/CIDR allowlist (comma-separated or repeated). When omitted, the existing setting is preserved; pass an empty value to clear restrictions")
	cobra.CheckErr(cmd.MarkFlagRequired("workspace-id"))
	return cmd
}
