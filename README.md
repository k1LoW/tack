# tack

`tack` is a minimal CLI for deploying a directory or a single file to [Tailor Platform](https://docs.tailor.tech/) [Static Website Hosting](https://docs.tailor.tech/guides/static-website-hosting.html).

> [!IMPORTANT]
> This is an **unofficial** and **experimental** tool.

## Features

- One-shot deploy: `tack up <path>` and the site is live
- Accepts a directory (uploaded recursively) or a single file
- Auto-creates the Static Website on first run, reuses it on subsequent runs
- Default Static Website name is `tack` so no extra config is needed
- Parallel chunked upload via the streaming `UploadFile` RPC
- Token auto-refresh via the Tailor SDK config (no separate login flow)
- Single binary with no external dependencies

## Install

Download binary from [releases page](https://github.com/k1LoW/tack/releases).

Or build from source:

```console
$ go install github.com/k1LoW/tack@latest
```

## Usage

If you have already logged in with `npx tailor-sdk login`, just run:

```console
$ tack up ./dist --workspace-id <WORKSPACE_ID>
```

`tack` reads tokens from the [Tailor SDK](https://github.com/tailor-platform/sdk) config (`~/.config/tailor-platform/config.yaml`). When the access token is expired it is refreshed automatically.

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

Each `tack up` is a **full replacement** of whatever was previously published. The Tailor Platform stores files per deployment and Publish flips an active-deployment pointer — there is no partial update, so every run uploads the complete site.

`.DS_Store`, `Thumbs.db`, and `desktop.ini` are skipped.

## Build

```console
$ go build .
```

## License

[MIT License](LICENSE)
