# Scanner adapter server

The server stub code is generated through swagger hub with `go-server` type based on the [swagger definitions](./api/swagger.yaml).

## Refactor

Move the following generated go files to a separate go package `mux` from the original package `spec` (renamed from pkg `go`):

- logger handler `logger.go`
- api handlers `api_scanner.go`
- mux routes `routers.go`

Move the server `main.go` to the `cmd/server` package.