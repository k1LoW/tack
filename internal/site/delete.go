package site

import (
	"context"
	"errors"
	"fmt"

	tailorv1 "buf.build/gen/go/tailor-inc/tailor/protocolbuffers/go/tailor/v1"
	"connectrpc.com/connect"
	tailorclient "github.com/k1LoW/tailor-client-go"
)

// DeleteOptions configures a Delete call.
type DeleteOptions struct {
	WorkspaceID string
	Name        string
}

// Delete removes a Static Website (and all of its deployments) from the
// workspace.
func Delete(ctx context.Context, client *tailorclient.Client, opts DeleteOptions) error {
	if opts.WorkspaceID == "" {
		return errors.New("workspace id is required")
	}
	if !nameRe.MatchString(opts.Name) {
		return fmt.Errorf("invalid static website name %q (must match %s)", opts.Name, nameRe.String())
	}

	if _, err := client.DeleteStaticWebsite(ctx, connect.NewRequest((&tailorv1.DeleteStaticWebsiteRequest_builder{
		WorkspaceId: opts.WorkspaceID,
		Name:        opts.Name,
	}).Build())); err != nil {
		if connect.CodeOf(err) == connect.CodeNotFound {
			return fmt.Errorf("static website %q not found in workspace %s", opts.Name, opts.WorkspaceID)
		}
		return fmt.Errorf("delete static website: %w", err)
	}
	return nil
}
