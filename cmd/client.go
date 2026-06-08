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

// newTailorClient builds a tailor-client-go Client. When either machine-user
// env var is set, it authenticates via the OAuth2 client_credentials grant;
// only when neither is set does it fall back to the SDK config
// (`~/.config/tailor-platform/config.yaml`). Forwarding a partial pair to
// tailor-client-go is deliberate: the library returns a clear "requires both
// clientID and clientSecret" error, which is the right signal when a CI job
// has set one variable but not the other (typo, mis-scoped secret, etc.). A
// silent fallback to whatever SDK token happens to be on disk would otherwise
// risk deploying as the wrong identity.
func newTailorClient(ctx context.Context) (*tailorclient.Client, error) {
	if id, secret := os.Getenv(envMachineUserClientID), os.Getenv(envMachineUserClientSecret); id != "" || secret != "" {
		return tailorclient.New(ctx, tailorclient.WithClientCredentials(id, secret))
	}
	return tailorclient.New(ctx)
}
