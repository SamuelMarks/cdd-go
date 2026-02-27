# Publishing Generated Client Libraries

This guide outlines the best practices for publishing the output of the `cdd_go` (language-to-OpenAPI / OpenAPI-to-language) tool. Specifically, it details how to manage a generated Go client library, keep it in sync with a server's OpenAPI spec, and publish it automatically.

## 1. The Git & Proxy Ecosystem

Just like the `cdd-go` library itself, Go client libraries are published via Git tags and cached by the global proxy (`proxy.golang.org`).

**A client library should ideally reside in its own Git repository** (e.g., `github.com/your-org/myapi-go-client`).
1. Your generator (`cdd_go from_openapi -i openapi.json -o .`) writes models and routes into the repo.
2. The code is committed, pushed, and tagged (`vX.Y.Z`).
3. Users fetch the client via `go get github.com/your-org/myapi-go-client`.

Docs for the client are automatically handled by `pkg.go.dev` once tagged.

## 2. Keeping the Client Library Up-To-Date (Automation)

To ensure your client library accurately reflects the active server logic, you should run `cdd_go` in a CI/CD pipeline (e.g., GitHub Actions) using a cron job. The pipeline will:
1. Download the latest `openapi.json` from your live API server.
2. Run `cdd_go` to generate the new AST/classes/routes.
3. Check for any Git differences.
4. If changes exist, commit the new code, tag a patch/minor release, and push.

### Example GitHub Action: `.github/workflows/update-client.yml`

Create this file in your client-library repository:

```yaml
name: Auto-update Go Client

on:
  schedule:
    # Run every night at midnight
    - cron: '0 0 * * *'
  workflow_dispatch: # Allow manual triggers

jobs:
  update-client:
    runs-on: ubuntu-latest
    steps:
      - name: Checkout Client Repository
        uses: actions/checkout@v3
        with:
          token: ${{ secrets.PAT_TOKEN }} # Use a PAT if pushing tags

      - name: Setup Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.21'

      - name: Install CDD Tool
        run: go install github.com/samuel/cdd-go/cmd/cdd_go@latest

      - name: Fetch Latest OpenAPI Spec
        run: |
          # Download the spec from your live backend server
          curl -sS https://api.yourdomain.com/openapi.json > openapi.json

      - name: Generate Go Client Code
        run: |
          # Generate models and interfaces directly into the client repo
          cdd_go from_openapi -i openapi.json -o pkg/
          
          # Clean up the spec file
          rm openapi.json

          # Tidy up Go modules
          go mod tidy

      - name: Check for Changes
        id: git-check
        run: |
          git diff --exit-code || echo "changes=true" >> $GITHUB_OUTPUT

      - name: Commit, Tag, and Push
        if: steps.git-check.outputs.changes == 'true'
        run: |
          git config user.name "github-actions[bot]"
          git config user.email "github-actions[bot]@users.noreply.github.com"
          
          git add .
          
          # Calculate a new tag (simple timestamp representation for automated patches, 
          # or integrate a semver bumping tool based on spec version)
          NEW_TAG="v1.0.$(date +'%Y%m%d%H%M')"
          
          git commit -m "chore: auto-update generated client from OpenAPI spec"
          git tag $NEW_TAG
          
          git push origin main
          git push origin $NEW_TAG
          
      - name: Trigger pkg.go.dev Indexing
        if: steps.git-check.outputs.changes == 'true'
        run: |
          # Tell the Go proxy to pull the newly generated client version
          GOPROXY=https://proxy.golang.org GO111MODULE=on go list -m github.com/${{ github.repository }}@$NEW_TAG
```

## 3. Local Static Docs for the Generated Client

If your developers require offline documentation or you want to host docs on an internal static portal:
1. Include `go install github.com/princjef/gomarkdoc/cmd/gomarkdoc@latest` in your CI pipeline.
2. Add a step to run `gomarkdoc --output docs/README.md ./pkg/...` 
3. Deploy the resulting `docs/` folder to GitHub Pages using the `actions/upload-pages-artifact` workflow.
