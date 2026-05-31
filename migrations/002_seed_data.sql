INSERT INTO product (id, sku, part_number, brand, category_id)
VALUES
  ('00000000-0000-0000-0000-000000000001', 'SKU-001', 'PART-001', 'WYZauto', '10000000-0000-0000-0000-000000000001'),
  ('00000000-0000-0000-0000-000000000002', 'SKU-002', 'PART-002', 'WYZauto', '10000000-0000-0000-0000-000000000001')
ON CONFLICT (id) DO NOTHING;

INSERT INTO translation (entity_type, entity_id, locale, field_name, field_value)
VALUES
  ('product', '00000000-0000-0000-0000-000000000001', 'en', 'name', 'Road Tyre'),
  ('product', '00000000-0000-0000-0000-000000000001', 'fr', 'name', 'Pneu Route'),
  ('product', '00000000-0000-0000-0000-000000000002', 'en', 'name', 'Sport Tyre'),
  ('product', '00000000-0000-0000-0000-000000000002', 'fr', 'name', 'Pneu Sport')
ON CONFLICT DO NOTHING;