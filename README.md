# Product Catalog Interview Challenge

A Go web application for managing a product catalog with variants, reviews, and inventory tracking.

Your task is to use your normal engineering workflow (including AI tools if you want) to identify, fix, and validate defects in this codebase.

## Goal

Deliver production-quality fixes with clear validation. Focus on correctness, not cosmetic refactors.

## Requirements

- Go 1.21+

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
   - **UI:** http://localhost:8080/
   - **Dashboard:** http://localhost:8080/stats
   - **API:** http://localhost:8080/products

The app will create a `catalog.db` SQLite database and seed it with sample data on first run.

**Tip:** Use a custom DB path so you can reset easily between runs:

```bash
DB_PATH=/tmp/product-catalog.db go run .
```

To reset the database, delete the file and restart.

## API Endpoints

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/products` | List all products (optional `?category=`) |
| `POST` | `/products` | Create a product |
| `GET` | `/products/:id` | Get a product |
| `PUT` | `/products/:id` | Update a product |
| `DELETE` | `/products/:id` | Soft-delete a product |
| `POST` | `/products/:id/purchase` | Purchase (decrement stock) |
| `GET` | `/products/:id/variants` | List variants for a product |
| `POST` | `/products/:id/variants` | Create a variant |
| `GET` | `/products/:id/variants/:vid` | Get a variant |
| `PUT` | `/products/:id/variants/:vid` | Update a variant |
| `DELETE` | `/products/:id/variants/:vid` | Delete a variant |
| `POST` | `/products/:id/variants/:vid/purchase` | Purchase a variant |
| `GET` | `/products/:id/inventory` | Variant inventory summary |
| `GET` | `/products/:id/reviews` | List reviews |
| `POST` | `/products/:id/reviews` | Create a review |
| `GET` | `/products/export` | Export products as CSV |
| `GET` | `/products/stats` | Catalog statistics |
| `GET` | `/search?q=` | Search products |
| `GET` | `/categories` | List categories |
| `GET` | `/sku/:sku` | Look up variant by SKU |
| `GET` | `/health` | Health check |

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
