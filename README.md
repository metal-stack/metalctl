# metalctl

*metalctl* is the command line client to access the [metal-api](https://github.com/metal-stack/metal-api).

## Installation

Download locations:

- [metalctl-linux-amd64](https://github.com/metal-stack/metalctl/releases/download/v0.8.3/metalctl-linux-amd64)
- [metalctl-darwin-amd64](https://github.com/metal-stack/metalctl/releases/download/v0.8.3/metalctl-darwin-amd64)
- [metalctl-windows-amd64](https://github.com/metal-stack/metalctl/releases/download/v0.8.3/metalctl-windows-amd64)

### Installation on Linux

```bash
curl -LO https://github.com/metal-stack/metalctl/releases/download/v0.8.3/metalctl-linux-amd64
chmod +x metalctl-linux-amd64
sudo mv metalctl-linux-amd64 /usr/local/bin/metalctl
```

### Installation on MacOS

```bash
curl -LO https://github.com/metal-stack/metalctl/releases/download/v0.8.3/metalctl-darwin-amd64
chmod +x metalctl-darwin-amd64
sudo mv metalctl-darwin-amd64 /usr/local/bin/metalctl
```

### Installation on Windows

```bash
curl -LO https://github.com/metal-stack/metalctl/releases/download/v0.8.3/metalctl-windows-amd64
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

## Available commands

Full documentation is generated out of the cobra command implementation with:

`metalctl markdown`

generated markdown is [here](docs/metalctl.md) and [here](https://docs.metal-stack.io/stable/external/metalctl/README/)
