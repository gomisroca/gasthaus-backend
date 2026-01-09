ALTER TABLE speisekarte
ADD COLUMN ingredients text[] NOT NULL DEFAULT '{}';
