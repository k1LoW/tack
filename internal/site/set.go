package site

import (
	"context"
	"errors"
	"fmt"

	tailorv1 "buf.build/gen/go/tailor-inc/tailor/protocolbuffers/go/tailor/v1"
	"connectrpc.com/connect"
	tailorclient "github.com/k1LoW/tailor-client-go"
)

// SetOptions configures a `tack set` invocation.
type SetOptions struct {
	WorkspaceID string
	Name        string
	// AllowedIPAddresses replaces the IP allowlist when AllowedIPAddressesSet
	// is true. Empty slice clears all restrictions.
	AllowedIPAddresses    []string
	AllowedIPAddressesSet bool
}

// Settings is the current state of a Static Website resource.
type Settings struct {
	Name               string   `json:"name"`
	WorkspaceID        string   `json:"workspace_id"`
	Description        string   `json:"description"`
	AllowedIPAddresses []string `json:"allowed_ip_addresses"`
	URL                string   `json:"url"`
}

// Set updates a Static Website resource's settings in place. The Static
// Website must already exist (call Deploy first to create it).
func Set(ctx context.Context, client *tailorclient.Client, opts SetOptions) (*Settings, error) {
	if opts.WorkspaceID == "" {
		return nil, errors.New("workspace id is required")
	}
	if !nameRe.MatchString(opts.Name) {
		return nil, fmt.Errorf("invalid static website name %q (must match %s)", opts.Name, nameRe.String())
	}
	if !opts.AllowedIPAddressesSet {
		return nil, errors.New("nothing to set: pass at least one settable flag")
	}

	resp, err := client.GetStaticWebsite(ctx, connect.NewRequest((&tailorv1.GetStaticWebsiteRequest_builder{
		WorkspaceId: opts.WorkspaceID,
		Name:        opts.Name,
	}).Build()))
	if err != nil {
		if connect.CodeOf(err) == connect.CodeNotFound {
			return nil, fmt.Errorf("static website %q not found in workspace %s", opts.Name, opts.WorkspaceID)
		}
		return nil, fmt.Errorf("get static website: %w", err)
	}
	current := resp.Msg.GetStaticwebsite()

	ws := (&tailorv1.StaticWebsite_builder{
		Name:               opts.Name,
		Description:        current.GetDescription(),
		AllowedIpAddresses: opts.AllowedIPAddresses,
	}).Build()
	updateResp, err := client.UpdateStaticWebsite(ctx, connect.NewRequest((&tailorv1.UpdateStaticWebsiteRequest_builder{
		WorkspaceId:   opts.WorkspaceID,
		Staticwebsite: ws,
	}).Build()))
	if err != nil {
		return nil, fmt.Errorf("update static website: %w", err)
	}
	updated := updateResp.Msg.GetStaticwebsite()

	return &Settings{
		Name:               updated.GetName(),
		WorkspaceID:        opts.WorkspaceID,
		Description:        updated.GetDescription(),
		AllowedIPAddresses: updated.GetAllowedIpAddresses(),
		URL:                updated.GetUrl(),
	}, nil
}
