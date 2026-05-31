## First Approach

- Keep the translation loading responsibility clear, testable, and efficient.

- Separate the product document building logic from the translation loading logic.

- ProductDocumentBuilder should only knows how to assemble the Elasticsearch document. It does not know whether translations come from PostgreSQL database or from cache. That responsibility belongs to the TranslationLoader interface.

- TranslationLoader is designed to bulk-load translations for multiple entity IDs in one database round trip. This avoids the N+1 query problem and makes the sync process more predictable as the number of products grows.

- For caching, I use an in-process TTL cache with entity-scoped keys. This means one updated product, attribute, or specification translation can be invalidated without flushing the whole cache. The cache is only an optimization; PostgreSQL remains the source of truth.

- For error handling, missing translations are not treated as fatal errors. The builder falls back to English when possible, and then to an empty string. Database and loader errors are returned with context so failures can be understood and debugged.

The main assumptions are documented in this README: 
    - cache TTL is configurable, 
    - cache invalidation is entity-level, 
    - missing translations should not fail indexing, 
    - delta sync can use updated_at as a cursor with additional safeguards for boundary cases.

## What I Would Improve Given More Time

- Add a real Elasticsearch indexing adapter behind an interface.
- Add a dependency resolver for attribute-level delta sync to find all affected product IDs efficiently.
- Use a composite delta cursor `(updated_at, id)` or a dedicated change-log table.
- Add request logging and structured metrics around cache hit rate and sync duration.
- Add more integration cases for missing translations and partial locale coverage.