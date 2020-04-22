# metalctl

*metalctl* is the command line client to access the [metal-api](https://github.com/metal-stack/metal-api).

## Installation

Download locations:

* [metalctl-linux-amd64](https://images.metal-pod.io/metalctl/metalctl-linux-amd64)
* [metalctl-darwin-amd64](https://images.metal-pod.io/metalctl/metalctl-darwin-amd64)
* [metalctl-windows-amd64](https://images.metal-pod.io/cloud-native/metalctl/metalctl-windows-amd64)

Via pre-build package:

```bash
sudo curl -fsSL https://images.metal-pod.io/metalctl/metalctl -o /usr/local/bin/metalctl
sudo chmod +x /usr/local/bin/metalctl
```

Self-build:

```bash
make
sudo ln -sf $(pwd)/bin/metalctl /usr/local/bin/metalctl
```

### Installation on Linux

```bash
sudo curl -fsSL https://images.metal-pod.io/metalctl/metalctl-linux-amd64 -o /usr/local/bin/metalctl
sudo chmod +x /usr/local/bin/metalctl
```

### Installation on MacOS

```bash
sudo curl -fsSL https://images.metal-pod.io/metalctl/metalctl-darvin-amd64 -o /usr/local/bin/metalctl
sudo chmod +x /usr/local/bin/metalctl
```

### Installation on Windows

```bash
curl -LO https://blobstore.fi-ts.io/cloud-native/metalctl/metalctl-windows-amd64
copy metalctl-windows-amd64 metalctl.exe
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
