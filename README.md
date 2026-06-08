# tack

`tack` is a one-shot deploy CLI for [Tailor Platform](https://docs.tailor.tech/) [Static Website Hosting](https://docs.tailor.tech/guides/static-website-hosting.html).

> [!IMPORTANT]
> This is an **unofficial** and **experimental** tool.

## Features

- One-shot deploy: `tack up <path>` and the site is live
- Accepts a directory (uploaded recursively) or a single file

## Install

**homebrew tap:**

```console
$ brew install k1LoW/tap/tack
```

**manually:**

Download binary from [releases page](https://github.com/k1LoW/tack/releases).

**go install:**

```console
$ go install github.com/k1LoW/tack@latest
```

## Usage

If you have already logged in with `npx tailor-sdk login`, just run:

```console
$ tack up ./dist --workspace-id <WORKSPACE_ID>
```

`tack` reads tokens from the [Tailor SDK](https://github.com/tailor-platform/sdk) config (`~/.config/tailor-platform/config.yaml`). When the access token is expired it is refreshed automatically.

### Authenticate as a Platform machine user (CI / headless)

For CI, batch jobs, and other headless environments, `tack` accepts Platform machine-user credentials via the same environment variables the Tailor SDK CLI uses for `tailor-sdk login --machineuser`:

```console
$ export TAILOR_PLATFORM_MACHINE_USER_CLIENT_ID=<client_id>
$ export TAILOR_PLATFORM_MACHINE_USER_CLIENT_SECRET=<client_secret>
$ tack up ./dist --workspace-id <WORKSPACE_ID>
```

When either variable is present, `tack` skips the SDK config entirely and fetches an access token via the OAuth2 `client_credentials` grant. A partial pair (only one of the two set, including the common CI pattern of an injected-but-empty secret) is forwarded as-is and fails fast with `requires both clientID and clientSecret`, instead of silently falling back to whatever SDK token happens to be on disk. The SDK-config flow above is used only when neither variable is present.

Deploy a single file:

```console
$ tack up resume.html --workspace-id <WORKSPACE_ID>
```

The file is served at `https://tack-<hash>.web.erp.dev/resume.html`.

Use a different Static Website name (multiple sites per workspace):

```console
$ tack up ./dist --workspace-id <WORKSPACE_ID> --name my-spa
```

Restrict access by IP / CIDR at deploy time:

```console
$ tack up ./dist --workspace-id <WORKSPACE_ID> \
    --allowed-ip-addresses 203.0.113.10/32,198.51.100.0/24
```

When `--allowed-ip-addresses` is omitted, any existing allowlist on the Static Website is preserved. Passing an empty value (`--allowed-ip-addresses=`) clears all restrictions.

Update settings without re-deploying:

```console
$ tack set --workspace-id <WORKSPACE_ID> \
    --allowed-ip-addresses 203.0.113.10/32

$ tack set --workspace-id <WORKSPACE_ID> --allowed-ip-addresses=
```

`tack set` requires the Static Website to already exist (run `tack up` first to create it).

List Static Websites:

```console
$ tack list --workspace-id <WORKSPACE_ID>
NAME     URL                              ALLOWED IPS
tack     https://tack-xxxx.web.erp.dev    -
my-spa   https://my-spa-yyyy.web.erp.dev  203.0.113.10/32
```

`tack ls` is the alias.

Delete a Static Website:

```console
$ tack delete --workspace-id <WORKSPACE_ID> --name my-spa
Delete static website "my-spa" in workspace ...? [y/N]: y
Deleted "my-spa"
```

Skip the confirmation with `-y` / `--yes`, e.g. for CI. `tack del` is the alias.

Machine-readable output for CI / scripting:

```console
$ tack up ./dist --workspace-id <WORKSPACE_ID> --output json
{
  "name": "tack",
  "workspace_id": "...",
  "deployment_id": "...",
  "url": "https://tack-xxxx.web.erp.dev",
  "files": 12
}
```

### Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--workspace-id` | | Tailor Platform workspace ID (required) |
| `--name` | `tack` | Static Website name (`^[a-z0-9][a-z0-9-]{1,61}[a-z0-9]$`) |
| `--allowed-ip-addresses` | | IP / CIDR allowlist (comma-separated or repeated). Omit to preserve, `--allowed-ip-addresses=` to clear |
| `-o`, `--output` | `text` | Output format: `text` or `json` |

## How it works

1. Resolves the access token from the Tailor SDK config (refreshing if expired)
2. Calls `GetStaticWebsite`. If the name does not exist yet, calls `CreateStaticWebsite`
3. Calls `CreateDeployment` to get a fresh deployment ID
4. Walks the target path, sends every file through the streaming `UploadFile` RPC in parallel
5. Calls `PublishDeployment` and prints the resulting URL

Each `tack up` is a **full replacement** of whatever was previously published. The Tailor Platform stores files per deployment and Publish flips an active-deployment pointer, so every run uploads the complete site.

`.DS_Store`, `Thumbs.db`, and `desktop.ini` are skipped.

## Build

```console
$ go build .
```

## License

[MIT License](LICENSE)
