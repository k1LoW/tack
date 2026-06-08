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
// env var is present, it authenticates via the OAuth2 client_credentials
// grant; only when neither is present does it fall back to the SDK config
// (`~/.config/tailor-platform/config.yaml`). Presence is checked with
// os.LookupEnv (not os.Getenv) so that a CI runner injecting a missing
// secret as an empty string still trips the machine-user branch — there,
// tailor-client-go emits a clear "requires both clientID and clientSecret"
// error instead of letting the helper silently fall through to whatever
// SDK token happens to be on disk, which would risk deploying as the
// wrong identity.
func newTailorClient(ctx context.Context) (*tailorclient.Client, error) {
	id, idSet := os.LookupEnv(envMachineUserClientID)
	secret, secretSet := os.LookupEnv(envMachineUserClientSecret)
	if idSet || secretSet {
		return tailorclient.New(ctx, tailorclient.WithClientCredentials(id, secret))
	}
	return tailorclient.New(ctx)
}
