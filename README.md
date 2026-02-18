# Product Catalog Interview Challenge

This repository is a small Go web application for managing products.

Your task is to use your normal engineering workflow (including AI tools if you want) to identify, fix, and validate defects in this codebase.

## Goal

Deliver production-quality fixes with clear validation. Focus on correctness, not cosmetic refactors.

## Requirements

- Go 1.21+
- CGO-enabled Go toolchain (required by `github.com/mattn/go-sqlite3`)

## Getting Started

1. Install dependencies:
   ```bash
   go mod download
   ```
2. Run the app:
   ```bash
   go run .
   ```
3. Open:
   - UI: `http://localhost:8080/`
   - API: `http://localhost:8080/products`

Optional: use a custom DB path so you can reset easily between runs.

```bash
DB_PATH=/tmp/product-catalog.db go run .
```

## What To Do

1. Find defects in behavior, reliability, and data handling.
2. Implement fixes with clean, maintainable code.
3. Validate each fix with automated checks and/or reproducible manual steps.
4. Document what you changed and why.

## Deliverables

Provide:

1. Code changes.
2. A short write-up covering:
   - defects found,
   - root cause for each defect,
   - fix summary,
   - validation evidence.
3. A brief note on AI usage:
   - which tools you used,
   - where AI helped,
   - where you disagreed with AI output and why.

## Evaluation Criteria

You will be evaluated on:

1. Correctness and completeness of fixes.
2. Validation quality (tests, reproducibility, edge cases).
3. Debugging depth and technical judgment.
4. Code quality and maintainability.
5. Effective and critical use of AI (not just accepting first output).
