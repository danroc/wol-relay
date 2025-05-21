# Wake-on-LAN Relay

Relay Wake-on-LAN (WOL) packets between multiple networks or interfaces. This
is useful for environments where devices are on separate subnets and need to be
woken up remotely.

## Features

- Listens for WOL packets on specified interfaces
- Relays valid WOL packets to all other monitored networks
- Prevents packet loops and duplicate broadcasts
- Lightweight and easy to deploy (single binary or Docker image)

## Usage

### Docker Compose

Add the following service to your `compose.yaml` to relay WOL packets between
the `eno1` and `eno2` network interfaces:

```yaml
services:
  wol-relay:
    image: ghcr.io/danroc/wol-relay:latest
    container_name: wol-relay
    network_mode: host
    command: eno1 eno2
    security_opt:
      - no-new-privileges:true
    restart: unless-stopped
```

> **Note:** `network_mode: host` is required for the container to access host
> network interfaces and broadcast packets.

### Standalone

You can also run the relay directly on your host:

```sh
# Build the binary
go build

# Run the relay
sudo ./wol-relay eno1 eno2
```

Replace `eno1` and `eno2` with the names of your network interfaces.

## Requirements

- Go 1.24+ (for building from source)
- Sufficient privileges to access network interfaces and send broadcast packets

## Development

- Run tests:

  ```sh
  go test ./...
  ```

- Build Docker image:

  ```sh
  docker build -t wol-relay .
  ```
