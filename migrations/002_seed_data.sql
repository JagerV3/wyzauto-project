```sql
INSERT INTO products (
    uuid,
    sku,
    created_at,
    updated_at
) VALUES (
    '00000000-0000-0000-0000-000000000001',
    'sample-product',
    NOW(),
    NOW()
)
ON CONFLICT (uuid) DO NOTHING;

-- Add translations / product document fields here
-- Example only. Adjust column names based on your schema.

INSERT INTO translations (
    product_uuid,
    locale,
    field_name,
    value,
    created_at,
    updated_at
) VALUES
(
    '00000000-0000-0000-0000-000000000001',
    'en',
    'productName',
    'Sample Product',
    NOW(),
    NOW()
)
ON CONFLICT DO NOTHING;
```
