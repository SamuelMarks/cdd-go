# Publishing

To publish the `cdd-go` compiler itself:

## Source Code / Library
For Go, publishing is as simple as tagging a Git release and pushing it to a public repository (e.g., GitHub). `go get` fetches it directly.

```bash
git tag v0.0.1
git push origin v0.0.1
```

Users can install via:
```bash
go get -u github.com/samuel/cdd-go
go install github.com/samuel/cdd-go/cmd/cdd-go@latest
```

## Documentation

To generate and serve static documentation:
```bash
make build_docs
```

To publish documentation to a popular Go ecosystem location, ensure the project is public on GitHub and visit `pkg.go.dev/github.com/samuel/cdd-go`. `pkg.go.dev` automatically indexes standard Go comments and structures.
