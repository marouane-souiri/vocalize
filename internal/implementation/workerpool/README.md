# Workerpool package

## Testing

### Test the Workerpool
```bash
# Run all Workerpool tests with verbose output
go test ./internal/discord -v -run "TestWorkerPool"
```

```bash
# Run a specific test
go test ./internal/discord -v -run "TestWorkerPool_BurstyLoad"
```

```bash
# Run a tests and skip test that need a lot of power
go test ./internal/discord -v -run "TestWorkerPool" -short
```

