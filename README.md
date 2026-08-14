# Concurrent Link Auditor

A bounded-concurrency Go library and command-line tool for checking HTTP and HTTPS endpoints. It preserves input order, emits human-readable or JSON reports, supports per-request timeouts, and exits non-zero when a link fails.

## Build and run

```bash
go build ./cmd/link-auditor
./link-auditor -workers 12 -timeout 8s https://example.com https://example.com/missing
```

Read URLs from a file and create a JSON report:

```bash
./link-auditor -input urls.txt -json > report.json
```

Blank lines and lines beginning with `#` are ignored.

## Library use

```go
results := auditor.Check(context.Background(), http.DefaultClient, urls, 8)
for _, result := range results {
    fmt.Println(result.URL, result.Healthy())
}
```

## Responsible use

Check only endpoints you are authorized to access. Keep worker counts reasonable and respect site terms, robots policies, and rate limits.

## Test

```bash
go test -race ./...
```

## License

MIT License.
