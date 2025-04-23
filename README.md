# port_finder

Simple ncurses program that calls `netstat -tunpl` to
find and display any connections given some defined
`PATTERN`

## Build

For a static build
```
go build -tags static port_finder
```
