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


## Assumptions

- Brand labels are translated through the translation table using the brand code as the entity identifier.
- Missing translations fall back to English, then empty string.
- Cache invalidation is entity-scoped.