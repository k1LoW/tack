/*
Copyright © 2026 Ken'ichiro Oyama <k1lowxb@gmail.com>
*/
package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/k1LoW/tack/internal/site"
)

func newDeleteCmd() *cobra.Command {
	var workspaceID string
	var name string
	var output string
	var yes bool

	cmd := &cobra.Command{
		Use:           "delete",
		Aliases:       []string{"del"},
		Short:         "Delete a Static Website (and all of its deployments)",
		Args:          cobra.NoArgs,
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			switch output {
			case "text", "json":
			default:
				return fmt.Errorf("invalid --output %q (must be text or json)", output)
			}

			if !yes {
				ok, err := confirm(cmd, fmt.Sprintf("Delete static website %q in workspace %s? [y/N]: ", name, workspaceID))
				if err != nil {
					return err
				}
				if !ok {
					return fmt.Errorf("aborted")
				}
			}

			ctx := cmd.Context()
			client, err := newTailorClient(ctx)
			if err != nil {
				return fmt.Errorf("init client: %w", err)
			}

			if err := site.Delete(ctx, client, site.DeleteOptions{
				WorkspaceID: workspaceID,
				Name:        name,
			}); err != nil {
				return err
			}

			switch output {
			case "text":
				fmt.Fprintf(cmd.OutOrStdout(), "Deleted %q\n", name)
			case "json":
				enc := json.NewEncoder(cmd.OutOrStdout())
				enc.SetIndent("", "  ")
				if err := enc.Encode(map[string]any{
					"name":         name,
					"workspace_id": workspaceID,
					"deleted":      true,
				}); err != nil {
					return fmt.Errorf("encode json: %w", err)
				}
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&workspaceID, "workspace-id", "", "Tailor Platform workspace id (uuid)")
	cmd.Flags().StringVar(&name, "name", defaultName, "static website name to delete")
	cmd.Flags().StringVarP(&output, "output", "o", "text", "output format: text or json")
	cmd.Flags().BoolVarP(&yes, "yes", "y", false, "skip confirmation prompt")
	cobra.CheckErr(cmd.MarkFlagRequired("workspace-id"))
	return cmd
}

func confirm(cmd *cobra.Command, prompt string) (bool, error) {
	fmt.Fprint(cmd.ErrOrStderr(), prompt)
	scanner := bufio.NewScanner(cmd.InOrStdin())
	if !scanner.Scan() {
		if err := scanner.Err(); err != nil {
			return false, err
		}
		return false, nil
	}
	answer := strings.ToLower(strings.TrimSpace(scanner.Text()))
	return answer == "y" || answer == "yes", nil
}
