# Kvetch

Kvetch is a small gRPC wrapper around the [Badger](https://github.com/dgraph-io/badger) key-value datastore

## Quickstart

Run the Kvetch container in Docker:

```bash
docker run --rm -v $PWD/data:/data -e DATASTORE=/data -p 7777:7777  syncromatics/kvetch:v0.5.1
```

Interact with Kvetch using `kvetchctl`:

```bash
GO111MODULE=on go get github.com/syncromatics/kvetch/cmd/kvetchctl@v0.5.1
export KVETCHCTL_ENDPOINT=localhost:7777 # host:port of Kvetch instance
kvetchctl set example/1 "first value"
kvetchctl set example/2 "second value"
kvetchctl set example/3 "third value"
kvetchctl get --prefix example/
```

It is also possible to run both `kvetch` and `kvetchctl` in the same docker container:

```bash
docker run -it --rm syncromatics/kvetch:v0.5.1 bash
DATASTORE=/data ./kvetch &
export KVETCHCTL_ENDPOINT=localhost:7777
./kvetchctl set example/1 "first value"
./kvetchctl set example/2 "second value"
./kvetchctl set example/3 "third value"
./kvetchctl get --prefix example/
```

More `kvetchctl` documentation is available in [docs/kvetchctl](docs/kvetchctl/kvetchctl.md)

## Configuration

Configuration is done via environmental variables. Refer to the tables below.

**General Settings**

| Name                        | Type     | Description                                               | Required | Default |
| --------------------------- | -------- | --------------------------------------------------------- | -------- | ------- |
| DATASTORE                   | string   | Directory where badger key data will be stored in.        | Yes      | `nil`   |
| GARBAGE_COLLECTION_INTERVAL | duration | Defines how often kvetch will attempt garbage collection. | No       | 5m      |
| PORT                        | int      | Port on which kvetch grpc service will run.               | No       | 7777    |
| PROMETHEUS_PORT             | int      | Port for use by Prometheus for metric gathering.          | No       | 80      |

**Optional BadgerDB Specific Settings** (More Detail @ https://github.com/dgraph-io/badger/blob/master/options.go)

Default values here are set to BadgerDB defaults and subject to change if package is updated. Refer to above link for more info.

| Name                                               | Type  | Description                                                  | Default |
| -------------------------------------------------- | ----- | ------------------------------------------------------------ | ------- |
| ENABLE_TRUNCATE                                    | bool  | Truncate indicates whether value log files should be truncated to delete corrupt data, if any. | False   |
| GARBAGE_COLLECTION_DISCARD_RATIO                   | float | Percentage of value log file that has to be expired or ready for garbage collection for that file to be eligible for garbage collection. | 0.5     |
| IN_MEMORY                                          | bool  | Sets InMemory mode to true. Everything is stored in memory. No value/sst files on disk are created. In case of a crash all data will be lost. | False   |
| LEVEL_ONE_SIZE                                     | int   | The maximum total size in bytes for Level 1 in the LSM.      | 20MB    |
| LEVEL_SIZE_MULTIPLIER                              | int   | Sets the ratio between the maximum sizes of contiguous levels in the LSM. Once a level grows to be larger than this ratio allowed, the compaction process will be triggered. | 10      |
| MAX_TABLE_SIZE                                     | int   | Sets the maximum size in bytes for each LSM table or file.   | 64MB    |
| NUMBER_OF_LEVEL_ZERO_TABLES                        | int   | Maximum number of Level 0 tables before compaction starts.   | 5       |
| NUMBER_OF_ZERO_LEVEL_TABLES_UNTIL_FORCE_COMPACTION | int   | Sets the number of Level 0 tables that once reached causes the DB to stall until compaction succeeds. | 10      |




## Building

[![build](https://github.com/syncromatics/kvetch/workflows/build/badge.svg)](https://github.com/syncromatics/kvetch/actions?query=workflow%3Abuild)
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
