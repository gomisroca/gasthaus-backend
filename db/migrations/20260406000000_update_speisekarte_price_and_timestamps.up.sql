ALTER TABLE speisekarte
  RENAME COLUMN price TO price_cents;

ALTER TABLE speisekarte
  ALTER COLUMN price_cents TYPE integer USING (price_cents * 100)::integer;

ALTER TABLE speisekarte
  ADD COLUMN IF NOT EXISTS created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW();