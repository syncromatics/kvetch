# Kvetch

Kvetch is a small gRPC wrapper around the [Badger](https://github.com/dgraph-io/badger) key-value datastore

## Quickstart

Run the Kvetch container in Docker:

```bash
docker run --rm -v $PWD/data:/data -e DATASTORE=/data -p 7777:7777  syncromatics/kvetch:0.4.0
```

Interact with Kvetch using `kvetchctl`:

```bash
export KVETCHCTL_ENDPOINT=localhost:7777 # host:port of Kvetch instance
kvetchctl set example/1 "first value"
kvetchctl set example/2 "second value"
kvetchctl set example/3 "third value"
kvetchctl get --prefix example/
```

More `kvetchctl` documentation is available in [docs/kvetchctl](docs/kvetchctl/kvetchctl.md)

## Building

![build](https://github.com/syncromatics/kvetch/workflows/build/badge.svg)
[![Docker Build Status](https://img.shields.io/docker/build/syncromatics/kvetch.svg)](https://hub.docker.com/r/syncromatics/kvetch/)

Building Kvetch requires the following:

- Docker
- [gogitver](https://github.com/syncromatics/gogitver)

To build and test the Kvetch Docker image alone:

```bash
make test
```

To build and package `kvetchctl`:

```bash
make package
```

To update generated code and documentation:

```bash
make generate
```

## Code of Conduct

We are committed to fostering an open and welcoming environment. Please read our [code of conduct](CODE_OF_CONDUCT.md) before participating in or contributing to this project.

## Contributing

We welcome contributions and collaboration on this project. Please read our [contributor's guide](CONTRIBUTING.md) to understand how best to work with us.

## License and Authors

[![GMV Syncromatics Engineering logo](https://secure.gravatar.com/avatar/645145afc5c0bc24ba24c3d86228ad39?size=16) GMV Syncromatics Engineering](https://github.com/syncromatics)

[![license](https://img.shields.io/github/license/syncromatics/kvetch.svg)](https://github.com/syncromatics/kvetch/blob/master/LICENSE)
[![GitHub contributors](https://img.shields.io/github/contributors/syncromatics/kvetch.svg)](https://github.com/syncromatics/kvetch/graphs/contributors)

This software is made available by GMV Syncromatics Engineering under the MIT license.
