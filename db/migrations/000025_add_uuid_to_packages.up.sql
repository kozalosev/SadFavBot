ALTER TABLE Packages ADD COLUMN IF NOT EXISTS unique_id uuid NOT NULL UNIQUE DEFAULT gen_random_uuid();
