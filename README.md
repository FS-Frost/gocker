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
gocker -help
gocker --help
```

Prints:

```shell
Repository: https://www.github.com/FS-Frost/gocker
Config: /home/aptus/.gocker/config.json

Usage:
a) With flags:
  -cmd string
        command to execute (default "bash")
  -config
        prints user config
  -container string
        container name
  -help
        prints usage
  -update
        updates gocker installation
  -version
        prints current version

  Example: 'gocker -container mysql -cmd "ls -l"'

b) Without flags:
  b1) Interactive shell:
    gocker

  b2) Custom command:
    gocker [container-name] [command] [arg1]...[argN]
      Example: 'gocker mysql ls -l'
```

### Example Output

```
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
