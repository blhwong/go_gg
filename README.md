# Go GG

- Track upset threads in fighting games using StartGG API
- Goroutine to poll StartGG for newest results
- Connects to websocket so that client gets latest updates


## Testing
```
go test ./...
```
with coverage
```
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```
