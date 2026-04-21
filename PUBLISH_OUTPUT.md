# Publishing Client Output

When you generate a client SDK using `cdd-go from_openapi to_sdk` or `to_sdk_cli`, a full Go module is scaffolded.

You can publish this new Go module separately on `pkg.go.dev` or an internal package registry like `Artifactory` by pushing the generated module to its own Git repository.

## GitHub Action Auto-Publish

To keep the client module in sync with the server API, configure a GitHub Action cronjob that polls the server's OpenAPI specification and regenerates the client module.

```yaml
name: "Update Client SDK"
on:
  schedule:
    - cron: '0 0 * * *' # Daily

jobs:
  sync:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: 1.25

      - name: Install cdd-go
        run: go install github.com/SamuelMarks/cdd-go/cmd/cdd-go@latest

      - name: Download server OpenAPI
        run: curl -sL "https://api.my-server.com/openapi.json" -o spec.json

      - name: Generate SDK
        run: cdd-go from_openapi to_sdk -i spec.json -o ./my-client-sdk

      - name: Commit and Push
        run: |
          git config user.name "github-actions[bot]"
          git config user.email "41898282+github-actions[bot]@users.noreply.github.com"
          git add ./my-client-sdk
          git commit -m "Auto-update Client SDK" || echo "No changes to commit"
          git push
```
