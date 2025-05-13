# Discord package

## Testing

### Test WS manager
```bash
# Run all WebSocket manager tests with verbose output
go test ./internal/discord -v -run "TestWSManager"
```

```bash
# Run a specific test
go test ./internal/discord -v -run "TestWSManager_Reconnect"
```
