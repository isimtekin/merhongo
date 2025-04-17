#!/bin/bash

# This script extracts test coverage data and updates README.md
# Run this after running "go test ./... -coverprofile=coverage.out"

# Ensure coverage.out exists
if [ ! -f coverage.out ]; then
    echo "Error: coverage.out not found. Run tests with coverage first."
    exit 1
fi

# Calculate total coverage
total_coverage=$(go tool cover -func=coverage.out | grep total | grep -Eo '[0-9]+\.[0-9]+')
total_coverage_int=${total_coverage%.*}

# Get coverage for each package
connection_coverage=$(go tool cover -func=coverage.out | grep connection | grep -Eo '[0-9]+\.[0-9]+' | head -1)
connection_coverage_int=${connection_coverage%.*}

model_coverage=$(go tool cover -func=coverage.out | grep model | grep -Eo '[0-9]+\.[0-9]+' | head -1)
model_coverage_int=${model_coverage%.*}

schema_coverage=$(go tool cover -func=coverage.out | grep schema | grep -Eo '[0-9]+\.[0-9]+' | head -1)
schema_coverage_int=${schema_coverage%.*}

query_coverage=$(go tool cover -func=coverage.out | grep query | grep -Eo '[0-9]+\.[0-9]+' | head -1)
query_coverage_int=${query_coverage%.*}

errors_coverage=$(go tool cover -func=coverage.out | grep errors | grep -Eo '[0-9]+\.[0-9]+' | head -1)
errors_coverage_int=${errors_coverage%.*}

# Create temporary file with updated coverage
awk -v conn="$connection_coverage_int" \
    -v model="$model_coverage_int" \
    -v schema="$schema_coverage_int" \
    -v query="$query_coverage_int" \
    -v errors="$errors_coverage_int" \
    -v total="$total_coverage_int" \
    '
    /\| connection  \| [0-9]+% *\|/ {
        sub(/\| connection  \| [0-9]+% *\|/, "| connection  | " conn "% |")
    }
    /\| model       \| [0-9]+% *\|/ {
        sub(/\| model       \| [0-9]+% *\|/, "| model       | " model "% |")
    }
    /\| schema      \| [0-9]+% *\|/ {
        sub(/\| schema      \| [0-9]+% *\|/, "| schema      | " schema "% |")
    }
    /\| query       \| [0-9]+% *\|/ {
        sub(/\| query       \| [0-9]+% *\|/, "| query       | " query "% |")
    }
    /\| errors      \| [0-9]+% *\|/ {
        sub(/\| errors      \| [0-9]+% *\|/, "| errors      | " errors "% |")
    }
    /\| \*\*Overall\*\* \| \*\*[0-9]+%\*\* *\|/ {
        sub(/\| \*\*Overall\*\* \| \*\*[0-9]+%\*\* *\|/, "| **Overall** | **" total "%** |")
    }
    /\[\!\[Test Coverage\]\(https:\/\/img.shields.io\/badge\/coverage-[0-9]+%25/ {
        sub(/\[\!\[Test Coverage\]\(https:\/\/img.shields.io\/badge\/coverage-[0-9]+%25/, "[![Test Coverage](https://img.shields.io/badge/coverage-" total "%25")
    }
    { print }
    ' README.md > README.md.tmp

# Replace README with updated version
mv README.md.tmp README.md

echo "README.md updated with coverage information:"
echo "  - Total coverage: ${total_coverage}%"
echo "  - Connection: ${connection_coverage}%"
echo "  - Model: ${model_coverage}%"
echo "  - Schema: ${schema_coverage}%"
echo "  - Query: ${query_coverage}%"
echo "  - Errors: ${errors_coverage}%"