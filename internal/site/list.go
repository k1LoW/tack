package site

import (
	"context"
	"errors"
	"fmt"

	tailorv1 "buf.build/gen/go/tailor-inc/tailor/protocolbuffers/go/tailor/v1"
	"connectrpc.com/connect"
	tailorclient "github.com/k1LoW/tailor-client-go"
)

const listPageSize = 100

// ListOptions configures a List call.
type ListOptions struct {
	WorkspaceID string
}

// List returns every Static Website registered in the given workspace.
// Results are fetched across all pages.
func List(ctx context.Context, client *tailorclient.Client, opts ListOptions) ([]*Settings, error) {
	if opts.WorkspaceID == "" {
		return nil, errors.New("workspace id is required")
	}

	var out []*Settings
	var pageToken string
	for {
		resp, err := client.ListStaticWebsites(ctx, connect.NewRequest((&tailorv1.ListStaticWebsitesRequest_builder{
			WorkspaceId: opts.WorkspaceID,
			PageToken:   pageToken,
			PageSize:    listPageSize,
		}).Build()))
		if err != nil {
			return nil, fmt.Errorf("list static websites: %w", err)
		}
		for _, w := range resp.Msg.GetStaticwebsites() {
			out = append(out, &Settings{
				Name:               w.GetName(),
				WorkspaceID:        opts.WorkspaceID,
				Description:        w.GetDescription(),
				AllowedIPAddresses: w.GetAllowedIpAddresses(),
				URL:                w.GetUrl(),
			})
		}
		pageToken = resp.Msg.GetNextPageToken()
		if pageToken == "" {
			break
		}
	}
	return out, nil
}
