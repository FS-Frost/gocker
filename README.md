# gocker

 CLI tool to execute commands on docker containers.

## Requirements

- [Go](https://go.dev/)
- [Docker](https://www.docker.com/)

## Install

```shell
go install github.com/FS-Frost/gocker@latest
```

## Usage

```shell
gocker
```

### Example Output

```shell
Containers:
1. golang
2. ubuntu
3. node

Enter container number or name: 2
Container: ubuntu (403c44416624)

Commands:
1. bash
2. sh
3. other

Enter command number or raw command to execute: 1
Command: bash

/usr/bin/docker exec -it ubuntu bash
root@403c44416624:/var/www/html#
```
