# Discord package

## Testing

### Test WS manager
```bash
# Run all WebSocket manager tests with verbose output
go test ./internal/discord/websocket -v -run "TestWSManager"
```

```bash
# Run a specific test
go test ./internal/discord/websocket -v -run "TestWSManager_Reconnect"
```

### Test Discord client
```bash
# Run all Client tests with verbose output
go test ./internal/discord/client -v -run "TestClient"
```

```bash
# Run a specific test
go test ./internal/discord/client -v -run "TestClient_Reconnect"
```

```bash
# Run Client tests with verbose output and skip the massive load test
go test ./internal/discord/client -v -run "TestClient" -short
```
