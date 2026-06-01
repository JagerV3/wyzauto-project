INSERT INTO product (id, sku, part_number, brand, category_id)
VALUES
  ('00000000-0000-0000-0000-000000000001', 'BP-OIL-5W30-1L', '5W30-1L', 'bosch', '00000000-0000-0000-0000-000000000010')
ON CONFLICT (id) DO NOTHING;

INSERT INTO attribute (id, code, metric_unit)
VALUES
  ('00000000-0000-0000-0000-000000000002', 'oil_grade', NULL)
ON CONFLICT (id) DO NOTHING;

INSERT INTO product_specification (id, product_id, attribute_id, value)
VALUES
  (
    '00000000-0000-0000-0000-000000000003',
    '00000000-0000-0000-0000-000000000001',
    '00000000-0000-0000-0000-000000000002',
    '5w30'
  )
ON CONFLICT (id) DO NOTHING;

INSERT INTO translation (entity_type, entity_id, locale, field_name, field_value, updated_at)
VALUES
  ('product', '00000000-0000-0000-0000-000000000001', 'en', 'productname', '5W-30 Engine Oil 1L', now()),
  ('product', '00000000-0000-0000-0000-000000000001', 'th', 'productname', 'น้ำมันเครื่อง 5W-30 1 ลิตร', now()),

  ('product', 'bosch', 'en', 'label', 'Bosch', now()),
  ('product', 'bosch', 'th', 'label', 'บอช', now()),

  ('attribute', '00000000-0000-0000-0000-000000000002', 'en', 'label', 'Oil Grade', now()),
  ('attribute', '00000000-0000-0000-0000-000000000002', 'th', 'label', 'เกรดน้ำมัน', now()),

  ('product_specification', '00000000-0000-0000-0000-000000000003', 'en', 'value_label', '5W-30', now()),
  ('product_specification', '00000000-0000-0000-0000-000000000003', 'th', 'value_label', '5W-30', now())
ON CONFLICT DO NOTHING;