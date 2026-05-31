INSERT INTO product (id, sku, name, brand, created_at, updated_at)
VALUES
  (1, 'SKU-001', 'Road Tyre', 'WYZauto', NOW(), NOW()),
  (2, 'SKU-002', 'Sport Tyre', 'WYZauto', NOW(), NOW())
ON CONFLICT (id) DO NOTHING;

INSERT INTO translation (id, product_id, locale, field_name, value, created_at, updated_at)
VALUES
  (1, 1, 'en', 'name', 'Road Tyre', NOW(), NOW()),
  (2, 1, 'fr', 'name', 'Pneu Route', NOW(), NOW()),
  (3, 2, 'en', 'name', 'Sport Tyre', NOW(), NOW()),
  (4, 2, 'fr', 'name', 'Pneu Sport', NOW(), NOW())
ON CONFLICT (id) DO NOTHING;