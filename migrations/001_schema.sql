CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE IF NOT EXISTS product (
    id uuid PRIMARY KEY,
    sku varchar NOT NULL,
    part_number varchar NOT NULL,
    brand varchar NOT NULL,
    category_id uuid NOT NULL
);

CREATE TABLE IF NOT EXISTS attribute (
    id uuid PRIMARY KEY,
    code varchar NOT NULL,
    metric_unit varchar NULL
);

CREATE TABLE IF NOT EXISTS product_specification (
    id uuid PRIMARY KEY,
    product_id uuid NOT NULL REFERENCES product(id),
    attribute_id uuid NOT NULL REFERENCES attribute(id),
    value varchar NOT NULL
);

CREATE TABLE IF NOT EXISTS translation (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    entity_type varchar NOT NULL,
    entity_id varchar NOT NULL,
    locale varchar NOT NULL,
    field_name varchar NOT NULL,
    field_value text NOT NULL,
    updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_translation_lookup
    ON translation (entity_type, entity_id, locale, field_name);

CREATE INDEX IF NOT EXISTS idx_translation_updated_at
    ON translation (updated_at);
