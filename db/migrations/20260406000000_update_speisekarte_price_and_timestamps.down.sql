ALTER TABLE speisekarte
  RENAME COLUMN price_cents TO price;

ALTER TABLE speisekarte
  ALTER COLUMN price TYPE float4 USING (price / 100.0)::float4;

ALTER TABLE speisekarte
  DROP COLUMN IF EXISTS created_at,
  DROP COLUMN IF EXISTS updated_at;