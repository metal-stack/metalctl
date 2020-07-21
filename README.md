# metalctl

*metalctl* is the command line client to access the [metal-api](https://github.com/metal-stack/metal-api).

## Installation

Download locations:

* [metalctl-linux-amd64](https://images.metal-stack.io/metalctl/metalctl-linux-amd64)
* [metalctl-darwin-amd64](https://images.metal-stack.io/metalctl/metalctl-darwin-amd64)
* [metalctl-windows-amd64](https://images.metal-stack.io/metalctl/metalctl-windows-amd64)

### Installation on Linux

```bash
sudo curl -fsSL https://images.metal-stack.io/metalctl/metalctl-linux-amd64 -o /usr/local/bin/metalctl
sudo chmod +x /usr/local/bin/metalctl
```

### Installation on MacOS

```bash
sudo curl -fsSL https://images.metal-stack.io/metalctl/metalctl-darwin-amd64 -o /usr/local/bin/metalctl
sudo chmod +x /usr/local/bin/metalctl
```

### Installation on Windows

```bash
curl -LO https://images.metal-stack.io/metalctl/metalctl-windows-amd64
copy metalctl-windows-amd64 metalctl.exe
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

Set up `metalctl` config, by first creating the config folder (`mkdir -p ~/.metalctl`), then set the `metalctl` URL within `~/.metalctl/config.yaml`:

```yaml
---
current: prod
contexts:
  prod:
    url: https://api.metal-stack.io/metal
    issuer_url: https://dex.metal-stack.io/dex
    client_id: metal_client
    client_secret: 456
```

## Available commands

Full documentation is generated out of the cobra command implementation with:

`metalctl markdown`

generated markdown is [here](docs/metalctl.md)
