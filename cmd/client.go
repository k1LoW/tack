/*
Copyright © 2026 Ken'ichiro Oyama <k1lowxb@gmail.com>
*/
package cmd

import (
	"context"
	"os"

	tailorclient "github.com/k1LoW/tailor-client-go"
)

// Env var names match the Tailor SDK CLI (`tailor-sdk login --machineuser`),
// so a machine user provisioned for the SDK can drive tack with no extra setup
// in CI / headless environments.
const (
	envMachineUserClientID     = "TAILOR_PLATFORM_MACHINE_USER_CLIENT_ID"
	envMachineUserClientSecret = "TAILOR_PLATFORM_MACHINE_USER_CLIENT_SECRET"
)

// newTailorClient builds a tailor-client-go Client. When both machine-user
// env vars are set, it authenticates via the OAuth2 client_credentials grant;
// otherwise it falls back to the SDK config (`~/.config/tailor-platform/config.yaml`).
func newTailorClient(ctx context.Context) (*tailorclient.Client, error) {
	if id, secret := os.Getenv(envMachineUserClientID), os.Getenv(envMachineUserClientSecret); id != "" && secret != "" {
		return tailorclient.New(ctx, tailorclient.WithClientCredentials(id, secret))
	}
	return tailorclient.New(ctx)
}
