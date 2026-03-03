#!/usr/bin/env bash
set -e

# Calculate test coverage
TEST_COV=$(go tool cover -func=coverage.out | grep total: | awk '{print $3}')
TEST_COV_URL="https://img.shields.io/badge/Test%20Coverage-${TEST_COV%?}%25-brightgreen.svg"

# For doc coverage, we run doc_cover.go
DOC_COV=$(go run ./scripts/doc_cover.go)
DOC_COV_URL="https://img.shields.io/badge/Doc%20Coverage-${DOC_COV%?}%25-brightgreen.svg"

# Update README.md using sed (cross-platform compatible fallback)
if [[ "$OSTYPE" == "darwin"* ]]; then
  sed -i '' "s|!\[Test Coverage\].*|![Test Coverage](${TEST_COV_URL})|" README.md
  sed -i '' "s|!\[Doc Coverage\].*|![Doc Coverage](${DOC_COV_URL})|" README.md
else
  sed -i "s|!\[Test Coverage\].*|![Test Coverage](${TEST_COV_URL})|" README.md
  sed -i "s|!\[Doc Coverage\].*|![Doc Coverage](${DOC_COV_URL})|" README.md
fi

git add README.md
