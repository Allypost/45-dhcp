# 45-DHCP

Simple implementation of a server which gives away IPs from a pool
and an example client that claims an IP.

## Requirements

[go](https://go.dev/) version 1.8+

## Usage

To build all the required components, run the following command

```bash
make all
```

To build individual components, there exist convenience commands for each component.

```bash
make server
make client
```

This should result in the appropriate binaries being created in the `bin` directory.

Usage of the binaries is described when run without any arguments.
