#!/usr/bin/env bash
set -e

# Run tests to generate coverage profile
echo "Running tests..."
go test -coverprofile=coverage.out ./...

# Calculate Test Coverage
TEST_COV=$(go tool cover -func=coverage.out | grep total: | grep -Eo '[0-9]+\.[0-9]+' || echo "0.0")

# Color Logic for Test Coverage
if (( $(echo "$TEST_COV >= 80" | bc -l) )); then
    TEST_COLOR="green"
elif (( $(echo "$TEST_COV >= 50" | bc -l) )); then
    TEST_COLOR="yellow"
else
    TEST_COLOR="red"
fi

# Calculate Doc Coverage
DOC_COV=$(go run scripts/doc_cover.go || echo "0.0")

# Color Logic for Doc Coverage
if (( $(echo "$DOC_COV >= 80" | bc -l) )); then
    DOC_COLOR="green"
elif (( $(echo "$DOC_COV >= 50" | bc -l) )); then
    DOC_COLOR="yellow"
else
    DOC_COLOR="red"
fi

echo "Test Coverage: ${TEST_COV}%"
echo "Doc Coverage: ${DOC_COV}%"

# Badges Markdown (Image only)
BADGES_MD="![Test Coverage](https://img.shields.io/badge/Test_Coverage-${TEST_COV}%25-${TEST_COLOR}) ![Doc Coverage](https://img.shields.io/badge/Doc_Coverage-${DOC_COV}%25-${DOC_COLOR})"

# Update README.md
if grep -q '!\[Test Coverage\]' README.md; then
    # Replace the existing badges line
    sed -i -E "s|^\!\[Test Coverage\].*|$BADGES_MD|" README.md
else
    # Insert badges directly below the first header
    awk -v badges="$BADGES_MD" '
        /^# cdd-go$/ {
            print
            print ""
            print badges
            next
        }
        {print}
    ' README.md > README.md.tmp && mv README.md.tmp README.md
fi

echo "README.md updated successfully."
