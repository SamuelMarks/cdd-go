# Publishing `cdd-go`

## Uploading to pkg.go.dev

Go modules are distributed via source code repositories (like GitHub). To publish a new version of `cdd-go`, tag the repository with a semantic version and push it. `pkg.go.dev` will automatically index it.

```bash
git tag v0.0.1
git push origin v0.0.1
# Trigger indexing
GOPROXY=https://proxy.golang.org GO111MODULE=on go list -m github.com/SamuelMarks/cdd-go@v0.0.1
```

## Publishing Docs

Run `make build_docs` to build a local folder of docs into `./docs/`.
You can upload this `./docs/` folder to GitHub Pages, an S3 bucket, or your own static server.

For example, to upload to S3:
```bash
aws s3 sync ./docs s3://my-cdd-docs-bucket --acl public-read
```

Go automatically generates library documentation for `pkg.go.dev` via code comments, so `pkg.go.dev/github.com/SamuelMarks/cdd-go` will serve as the most popular location for standard API reference.
