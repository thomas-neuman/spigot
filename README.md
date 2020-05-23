![Go](https://github.com/thomas-neuman/spigot/workflows/Go/badge.svg?branch=master)

## NOTE: Spigot is currently in early-stage development
The security model is largely unassessed, and things may very well break.

# Spigot

Spigot is a VPN (Virtual Private Network) client backed by NKN.
It enables you to construct a basic layer 2 connection among a set
of peers over any underlying fabric, as long as the Internet is
accessible.

## Motivation

NKN presents an interesting opportunity to tunnel network traffic
across geopolitical boundaries, or even just simply through NAT
without needing to open up one's own public IP. Solutions currently
exist, with additional ones being built, but so far all seem to rely
too heavily on interacting with NKN's model of identity (read:
addressing), or are primarily built for point-to-point tunnels.
Therefore, there is a gap to be bridged: to provide the same
connectivity benefits to users, without needing to adapt one's
application or workflows to NKN's schema. Hence, Spigot exposes
each peer as a private IPv4 address, while still maintaining the
same authorization guarantees granted by the underlying NKN peer
structure.

## Usage

### Building

The simplest way to build the project:
```sh
make
```

Run Spigot from under the `dist` directory with
```
./spigot
```

Spigot runs in the foreground. Note that, because Spigot creates
a TAP interface on your host, it may need to run with elevated
provileges.

### Configuring

Spigot is configured by JSON file stored at `/etc/spigot/config.json`.
This provides the main configuration options:

* `ip_address`: IPv4 address to be granted to this peer.
* `interface_name`: Name of the TAP interface to be created.
* `private_seed_file`: File containing the hexadecimal seed for your peer.
* `authorized_keys`: List of public keys which are allowed to communicate.
* `static_routes`: List of structures defining routes to peers.

Each static route is structured as:

* `destination`: IPv4 address of the peer.
* `nexthop`: NKN address of the peer.

See also the contents of the `example` directory, for a sample system
configuration. This also includes a SystemD service definition file,
in order to run the Spigot process in the background at init.

## See also

[NKN](https://www.nkn.org/)
[NKN Tunnel](https://github.com/nknorg/nkn-tunnel)
