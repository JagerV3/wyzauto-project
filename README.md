# wyzauto-project

A small Go implementation of a translation loading layer for WYZauto's product sync pipeline.

The exercise focuses on query efficiency, interface design, cache strategy, error handling, testability, and clear reasoning rather than line count.

## Tech Stack

- Go 1.22+
- net/http
- PostgreSQL
- pgxpool
- Docker Compose
- Makefile
- golangci-lint

## Problem Summary

WYZauto stores product catalogue data in PostgreSQL and builds buyer-facing search documents for Elasticsearch. Product content is multi-locale, and translations for product names, descriptions, attribute labels, and specification value labels are stored in a generic `translation` table.

The current bottleneck is the N+1 query pattern: translations are loaded one-by-one per product or entity. This implementation introduces a `TranslationLoader` that can bulk-load translations for multiple entity IDs in one database round trip, apply locale filtering, cache results with TTL, and return graceful fallbacks for missing translations.

## Package Structure

```text
cmd/server                HTTP entrypoint
internal/config           Environment-based configuration
internal/domain           Core structs and constants
internal/handler          HTTP handlers
internal/repository       PostgreSQL repositories using pgxpool
internal/service          Translation loader, cache, and document builder
migrations                PostgreSQL schema
```

## Design Overview

```text
Product Sync / HTTP Request
        |
        v
ProductDocumentBuilder
        |
        v
TranslationLoader Interface
        |
        +--> In-process TTL Cache
        |
        +--> PostgreSQL Translation Repository
        |
        v
Structured TranslationMap
        |
        v
ProductDocument struct
        |
        v
Elasticsearch document shape
```

The `ProductDocumentBuilder` only knows how to assemble the Elasticsearch document. It does not know whether translations come from cache or PostgreSQL. Translation retrieval is hidden behind the `TranslationLoader` interface.

## Query Efficiency

The `TranslationLoader` is responsible for retrieving translations for multiple entities in a single database query. By batching entity IDs into one request, the loader avoids the N+1 query pattern and keeps sync performance predictable as product volume grows.

Example query strategy:

```sql
SELECT entity_type, entity_id, locale, field_name, field_value, updated_at
FROM translation
WHERE entity_type = $1
  AND entity_id = ANY($2)
  AND locale = ANY($3)
ORDER BY entity_id, locale, field_name;
```

## Interface Design

The loader is defined as an interface so it can be tested in isolation and mocked by the document builder tests.

```go
type TranslationLoader interface {
    Load(ctx context.Context, entityType domain.EntityType, entityIDs []string, locales []string) (domain.TranslationMap, error)
    Invalidate(entityType domain.EntityType, entityID string)
    LoadUpdatedSince(ctx context.Context, cursor time.Time, locales []string) ([]domain.Translation, error)
}
```

Concrete database details are kept inside the repository package.

## Cache Strategy

For caching, this project uses an in-process TTL cache with entity-scoped keys such as product ID, attribute ID, and product specification ID. This allows one updated translation to invalidate only the affected entity instead of flushing the entire cache.

The cache is protected by a mutex to keep reads and writes thread-safe during concurrent sync operations. The cache is only an optimization; PostgreSQL remains the source of truth.

## Error Handling

Missing translations are not treated as fatal errors. The document builder falls back to English when possible, and then to an empty string if no translation exists. This prevents one incomplete translation from failing the whole indexing process.

Database and loader errors are returned with additional context using Go error wrapping so failures can be traced and debugged more easily.

## Delta Sync Design

To support delta sync, I would extend the translation loading flow with a cursor-based method that loads only translations updated after the last successful sync timestamp.

The loader exposes a method such as `LoadUpdatedSince(ctx, cursor, locales)` which queries the `translation` table using `updated_at` as the main filter. The result is grouped by `entity_type` and `entity_id`, so the sync process can identify which cache entries and Elasticsearch documents are affected.

Example query strategy:

```sql
SELECT entity_type, entity_id, locale, field_name, field_value, updated_at
FROM translation
WHERE updated_at > $1
  AND locale = ANY($2)
ORDER BY updated_at ASC;
```

After loading changed rows, the sync process should invalidate cache entries only for affected entities instead of flushing the whole cache. The product sync pipeline can then rebuild only impacted product documents.

Impact scope depends on the changed entity type:

- Product translation changed: rebuild that product document only.
- Product specification value label changed: rebuild the product that owns that specification.
- Attribute label changed: rebuild products that use that attribute.

The main trade-off is simplicity versus correctness. A timestamp cursor is easy to implement and efficient for this schema, but it can miss rows when multiple updates share the same `updated_at` value around the cursor boundary. To reduce this risk, I would either store a composite cursor such as `(updated_at, translation_id)` if an ID is available, or use a small overlap window and make re-indexing idempotent. Given more time, I would consider a change-log table or CDC-based approach for stronger reliability.

## Assumptions

- `translation.entity_id` is stored as a string because it may represent product UUIDs, attribute UUIDs, specification UUIDs, or brand codes.
- Brand labels are treated as `entity_type = product` with `entity_id = brand code` for this exercise implementation.
- Missing requested locale values fallback to English, then to an empty string.
- Cache TTL is configurable through `CACHE_TTL_SECONDS`.
- Cache invalidation is entity-scoped.
- PostgreSQL is the source of truth; cache is only an optimization.
- Delta sync uses `updated_at` as a simple cursor, with documented production safeguards.

## Setup

```bash
git clone <your-repo-url>
cd wyzauto-project
make setup
```

Start PostgreSQL:

```bash
docker compose up -d
```

Run the HTTP server:

```bash
make run
```

Health check:

```bash
curl http://localhost:8080/health
```

## Seed Test Data

The product document endpoint requires product data to exist in PostgreSQL.

For local testing, insert the sample SQL data manually before calling the document endpoint.

Example:

```bash
psql <your-database-url> -f scripts/seed.sql
```

Or connect to PostgreSQL and run your insert SQL manually.

## Build Product Document

After the sample product data exists in the database:

```bash
curl http://localhost:8080/products/00000000-0000-0000-0000-000000000001/document | jq
```
