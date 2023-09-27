# metalctl

*metalctl* is the command line client to access the [metal-api](https://github.com/metal-stack/metal-api).

## Installation

Download locations:

- [metalctl-linux-amd64](https://github.com/metal-stack/metalctl/releases/latest/download/metalctl-linux-amd64)
- [metalctl-darwin-amd64](https://github.com/metal-stack/metalctl/releases/latest/download/metalctl-darwin-amd64)
- [metalctl-darwin-arm64](https://github.com/metal-stack/metalctl/releases/latest/download/metalctl-darwin-arm64)
- [metalctl-windows-amd64](https://github.com/metal-stack/metalctl/releases/latest/download/metalctl-windows-amd64)

### Installation on Linux

```bash
curl -LO https://github.com/metal-stack/metalctl/releases/latest/download/metalctl-linux-amd64
chmod +x metalctl-linux-amd64
sudo mv metalctl-linux-amd64 /usr/local/bin/metalctl
```

### Installation on MacOS

For x86 based Macs:

```bash
curl -LO https://github.com/metal-stack/metalctl/releases/latest/download/metalctl-darwin-amd64
chmod +x metalctl-darwin-amd64
sudo mv metalctl-darwin-amd64 /usr/local/bin/metalctl
```

For Apple Silicon (M1) based Macs:

```bash
curl -LO https://github.com/metal-stack/metalctl/releases/latest/download/metalctl-darwin-arm64
chmod +x metalctl-darwin-arm64
sudo mv metalctl-darwin-arm64 /usr/local/bin/metalctl
```

### Installation on Windows

```bash
curl -LO https://github.com/metal-stack/metalctl/releases/latest/download/metalctl-windows-amd64
copy metalctl-windows-amd64 metalctl.exe
```

### metalctl update

In order to keep your local `metalctl` installation up to date, you can update the binary like this:

```bash
metalctl update check
latest version:v0.8.3 from:2020-08-13T11:55:14Z
local  version:v0.8.2 from:2020-08-12T09:27:39Z
metalctl is not up to date

metalctl update do
# a download with progress bar starts and replaces the binary. If the binary has root permissions please execute
sudo metalctl update do
# instead
```

### Built from project

```bash
make
sudo ln -sf $(pwd)/bin/metalctl /usr/local/bin/metalctl
```

## Configuration

Set up auto-completion for `metalctl`, e.g. add to your `~/.bashrc`:

```bash
source <(metalctl completion bash)
```

Set up `metalctl` config, by first creating the config folder (`mkdir -p ~/.metalctl`), then set the values according to your installation in `~/.metalctl/config.yaml`:

```yaml
---
current: prod
contexts:
  prod:
    url: https://api.metal-stack.io/metal
    issuer_url: https://dex.metal-stack.io/dex
    client_id: metal_client
    client_secret: 456
    hmac: YOUR_HMAC
```

Optional you can specify `issuer_type: generic` if you use other issuers as Dex, e.g. Keycloak (this will request scopes `openid,profile,email`):

```bash
contexts:
  prod:
    url: https://api.metal-stack.io/metal
    issuer_url: https://keycloak.somedomain.io
    issuer_type: generic
    client_id: my-client-id
    client_secret: my-secret
```

If you must specify special scopes for your issuer, you can use `custom_scopes`:

```bash
contexts:
  prod:
    url: https://api.metal-stack.io/metal
    issuer_url: https://keycloak.somedomain.io
    custom_scopes: roles,openid,profile,email
    client_id: my-client-id
    client_secret: my-secret
```

## Available commands

Full documentation is generated out of the cobra command implementation with:

`metalctl markdown`

generated markdown is [here](docs/metalctl.md) and [here](https://docs.metal-stack.io/stable/external/metalctl/README/)

## Development

For MacOS users, running the tests might throw an error because tests are utilizing the [] library in order to manipulate the `time.Now` function. The patch allows testing with fixed timestamps.

Instead, MacOS users can utilize the `make test-in-docker` target to execute the tests.
