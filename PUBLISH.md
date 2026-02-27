# Publishing the `cdd-go` Library & CLI

This guide explains how to publish the `cdd-go` Go module to the global Go ecosystem, how to generate local static documentation, and how to publish your documentation to the standard Go registry.

## 1. Publishing to the Go Ecosystem

Unlike Node.js (npm) or Rust (crates.io), Go does not use a centralized upload registry. Instead, the Go ecosystem relies on Version Control Systems (VCS) and the public Go Module Mirror (`proxy.golang.org`). 

To publish or update your Go module:

1. **Commit your changes** to your main branch.
2. **Tag a semantic version** (e.g., `v1.0.0`):
   ```bash
   git tag v1.0.0
   git push origin v1.0.0
   ```
3. **Trigger the Go Proxy to cache your module**:
   Run the following command to force the official proxy to cache your new release:
   ```bash
   GOPROXY=https://proxy.golang.org GO111MODULE=on go list -m github.com/samuel/cdd-go@v1.0.0
   ```
   *Your module is now published! Anyone can install it via `go get github.com/samuel/cdd-go@v1.0.0`.*

## 2. Generating Local Docs for Static Serving

Go's official tool to view documentation locally is `pkgsite`, but it runs an active HTTP server. If you need to generate a **local folder of static files** (e.g., Markdown or HTML) for serving on a static host (like GitHub Pages, Netlify, or an S3 bucket), the most popular community tool is `gomarkdoc`.

**Step A: Install `gomarkdoc`**
```bash
go install github.com/princjef/gomarkdoc/cmd/gomarkdoc@latest
```

**Step B: Generate a static `/docs` folder**
```bash
mkdir -p docs
gomarkdoc --output docs/README.md ./...
```
*(This traverses your Go code, reads all the comments we carefully preserved in the AST, and dumps beautifully formatted static Markdown files into the `docs/` folder, which can be deployed to any static host).*

**Alternative: Running standard `pkgsite` locally**
If you prefer to browse it identically to how it looks on the official Go registry:
```bash
go install golang.org/x/pkgsite/cmd/pkgsite@latest
pkgsite -http=:8080
# Visit http://localhost:8080/github.com/samuel/cdd-go
```

## 3. Publishing Docs to the Most Popular Location

The standard, most popular location for Go documentation is **[pkg.go.dev](https://pkg.go.dev/)**.

You do not need to manually upload HTML files to `pkg.go.dev`. It automatically indexes your public GitHub repository as soon as the module proxy fetches your tag.

**To ensure your docs are instantly available:**
1. Follow the exact steps in Part 1 to tag and push your code.
2. Visit `https://pkg.go.dev/github.com/samuel/cdd-go@v1.0.0`.
3. The site will fetch your code, parse the exported types/functions and docstrings, and generate standard documentation pages automatically.
