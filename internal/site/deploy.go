package site

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"regexp"
	"time"

	tailorv1 "buf.build/gen/go/tailor-inc/tailor/protocolbuffers/go/tailor/v1"
	"connectrpc.com/connect"
	tailorclient "github.com/k1LoW/tailor-client-go"
	"golang.org/x/sync/errgroup"
)

const (
	overallTimeout = 10 * time.Minute
	uploadTimeout  = 2 * time.Minute
	maxConcurrency = 5
)

// nameRe matches the static website name validation enforced by Tailor Platform.
var nameRe = regexp.MustCompile(`^[a-z0-9][a-z0-9-]{1,61}[a-z0-9]$`)

// DeployOptions configures one Deploy run.
type DeployOptions struct {
	WorkspaceID string
	Name        string
	Path        string
	// AllowedIPAddresses, when AllowedIPAddressesSet is true, replaces the
	// static website's IP allowlist (empty slice clears all restrictions).
	// When false, the existing allowlist is left untouched.
	AllowedIPAddresses    []string
	AllowedIPAddressesSet bool
	// Progress receives human-readable status lines. When nil, progress is
	// discarded. Set to os.Stdout for an interactive run.
	Progress io.Writer
}

// DeployResult summarises a completed deploy.
type DeployResult struct {
	Name         string `json:"name"`
	WorkspaceID  string `json:"workspace_id"`
	DeploymentID string `json:"deployment_id"`
	URL          string `json:"url"`
	Files        int    `json:"files"`
}

// Deploy ensures the static website exists, creates a new deployment, uploads
// every file under Path, and publishes the deployment. Each call is a full
// replacement of whatever was previously published.
func Deploy(ctx context.Context, client *tailorclient.Client, opts DeployOptions) (*DeployResult, error) {
	if opts.WorkspaceID == "" {
		return nil, errors.New("workspace id is required")
	}
	if !nameRe.MatchString(opts.Name) {
		return nil, fmt.Errorf("invalid static website name %q (must match %s)", opts.Name, nameRe.String())
	}

	progress := opts.Progress
	if progress == nil {
		progress = io.Discard
	}

	entries, err := Collect(opts.Path)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(ctx, overallTimeout)
	defer cancel()

	if err := ensureStaticWebsite(ctx, client, opts); err != nil {
		return nil, err
	}

	fmt.Fprintf(progress, "Creating deployment for %q...\n", opts.Name)
	createResp, err := client.CreateDeployment(ctx, connect.NewRequest((&tailorv1.CreateDeploymentRequest_builder{
		WorkspaceId: opts.WorkspaceID,
		Name:        opts.Name,
	}).Build()))
	if err != nil {
		return nil, fmt.Errorf("create deployment: %w", err)
	}
	deploymentID := createResp.Msg.GetDeploymentId()

	fmt.Fprintf(progress, "Uploading %d file(s)...\n", len(entries))
	if err := uploadAll(ctx, client, opts.WorkspaceID, deploymentID, entries); err != nil {
		return nil, err
	}

	fmt.Fprintf(progress, "Publishing deployment...\n")
	pubResp, err := client.PublishDeployment(ctx, connect.NewRequest((&tailorv1.PublishDeploymentRequest_builder{
		WorkspaceId:  opts.WorkspaceID,
		DeploymentId: deploymentID,
	}).Build()))
	if err != nil {
		return nil, fmt.Errorf("publish deployment: %w", err)
	}

	return &DeployResult{
		Name:         opts.Name,
		WorkspaceID:  opts.WorkspaceID,
		DeploymentID: deploymentID,
		URL:          pubResp.Msg.GetUrl(),
		Files:        len(entries),
	}, nil
}

func ensureStaticWebsite(ctx context.Context, client *tailorclient.Client, opts DeployOptions) error {
	resp, err := client.GetStaticWebsite(ctx, connect.NewRequest((&tailorv1.GetStaticWebsiteRequest_builder{
		WorkspaceId: opts.WorkspaceID,
		Name:        opts.Name,
	}).Build()))
	if err == nil {
		if !opts.AllowedIPAddressesSet {
			return nil
		}
		current := resp.Msg.GetStaticwebsite()
		ws := (&tailorv1.StaticWebsite_builder{
			Name:               opts.Name,
			Description:        current.GetDescription(),
			AllowedIpAddresses: opts.AllowedIPAddresses,
		}).Build()
		if _, err := client.UpdateStaticWebsite(ctx, connect.NewRequest((&tailorv1.UpdateStaticWebsiteRequest_builder{
			WorkspaceId:   opts.WorkspaceID,
			Staticwebsite: ws,
		}).Build())); err != nil {
			return fmt.Errorf("update static website: %w", err)
		}
		return nil
	}
	if connect.CodeOf(err) != connect.CodeNotFound {
		return fmt.Errorf("get static website: %w", err)
	}

	ws := (&tailorv1.StaticWebsite_builder{
		Name:               opts.Name,
		AllowedIpAddresses: opts.AllowedIPAddresses,
	}).Build()
	if _, err := client.CreateStaticWebsite(ctx, connect.NewRequest((&tailorv1.CreateStaticWebsiteRequest_builder{
		WorkspaceId:   opts.WorkspaceID,
		Staticwebsite: ws,
	}).Build())); err != nil {
		return fmt.Errorf("create static website: %w", err)
	}
	return nil
}

func uploadAll(ctx context.Context, client *tailorclient.Client, workspaceID, deploymentID string, entries []Entry) error {
	eg, ctx := errgroup.WithContext(ctx)
	eg.SetLimit(maxConcurrency)
	for _, e := range entries {
		eg.Go(func() error {
			return uploadOne(ctx, client, workspaceID, deploymentID, e)
		})
	}
	return eg.Wait()
}

func uploadOne(ctx context.Context, client *tailorclient.Client, workspaceID, deploymentID string, e Entry) error {
	ctx, cancel := context.WithTimeout(ctx, uploadTimeout)
	defer cancel()
	f, err := os.Open(e.AbsPath)
	if err != nil {
		return fmt.Errorf("open %s: %w", e.AbsPath, err)
	}
	defer f.Close()
	if err := client.UploadFileFromReader(ctx, tailorclient.UploadFileParams{
		WorkspaceID:  workspaceID,
		DeploymentID: deploymentID,
		FilePath:     e.RelPath,
		ContentType:  e.ContentType,
	}, f); err != nil {
		return fmt.Errorf("upload %s: %w", e.RelPath, err)
	}
	return nil
}
