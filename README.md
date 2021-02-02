# MUD
Mult-user dungeon written in Go

## Start server

Must be in same directory as `world.db` file

```bash
go install
mud
```
or
```bash
go build && ./mud
```
or
```bash
go run .
```

## Connecting

Uses TCP connection

Default port is 9001

Host is localhost for same machine


e.g. telnet:
```bash
telnet <HOST> <PORT>
```