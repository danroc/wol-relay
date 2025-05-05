# Wake-on-LAN Relay

Relay Wake-on-LAN packets between networks.

## Usage

In a Docker compose file, you can add the following service to relay WOL
packets between the `eno1` and `eno2` network interfaces:

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

  # Other services...
```
