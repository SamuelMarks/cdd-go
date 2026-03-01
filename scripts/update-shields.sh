#!/usr/bin/env bash
set -e

# Run tests to generate coverage profile
echo "Running tests..."
# We run go test -cover to get the percentage per package
# and average it out, since `go tool cover` sometimes fails on main packages in Go 1.25.
TEST_OUT=$(go test -cover ./... || echo "0.0")

# Calculate Test Coverage
TEST_COV=$(echo "$TEST_OUT" | grep "coverage:" | sed -n 's/.*coverage: \([0-9.]*\)%.*/\1/p' | awk '{sum+=$1; n++} END {if (n>0) printf "%.1f", sum/n; else print "0.0"}')

# Color Logic for Test Coverage
if (( $(echo "$TEST_COV >= 80" | bc -l) )); then
    TEST_COLOR="brightgreen"
elif (( $(echo "$TEST_COV >= 50" | bc -l) )); then
    TEST_COLOR="yellow"
else
    TEST_COLOR="red"
fi

# Calculate Doc Coverage
DOC_COV=$(go run scripts/doc_cover.go || echo "0.0")

# Color Logic for Doc Coverage
if (( $(echo "$DOC_COV >= 80" | bc -l) )); then
    DOC_COLOR="brightgreen"
elif (( $(echo "$DOC_COV >= 50" | bc -l) )); then
    DOC_COLOR="yellow"
else
    DOC_COLOR="red"
fi

echo "Test Coverage: ${TEST_COV}%"
echo "Doc Coverage: ${DOC_COV}%"

# Badges Markdown
TEST_BADGE="[![Test Coverage](https://img.shields.io/badge/test_coverage-${TEST_COV}%25-${TEST_COLOR}.svg)](#)"
DOC_BADGE="[![Doc Coverage](https://img.shields.io/badge/doc_coverage-${DOC_COV}%25-${DOC_COLOR}.svg)](#)"

# Update README.md
if grep -q '\[!\[Test Coverage\]' README.md; then
    # Replace the existing badges line
    sed -i -E "s|\[\!\[Test Coverage\].*|$TEST_BADGE|" README.md
fi

if grep -q '\[!\[Doc Coverage\]' README.md; then
    # Replace the existing badges line
    sed -i -E "s|\[\!\[Doc Coverage\].*|$DOC_BADGE|" README.md
fi

echo "README.md updated successfully."
