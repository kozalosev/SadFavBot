CREATE TABLE IF NOT EXISTS Locations (
    id serial PRIMARY KEY,
    latitude double precision NOT NULL,
    longitude double precision NOT NULL
);

ALTER TABLE Favs ADD COLUMN IF NOT EXISTS location_id integer CONSTRAINT fk_favs_location REFERENCES Locations(id);

ALTER TABLE Favs DROP CONSTRAINT IF EXISTS ck_favs_content;
ALTER TABLE Favs ADD CONSTRAINT ck_favs_content CHECK ( file_id IS NOT NULL AND file_unique_id IS NOT NULL
                                                         OR text_id IS NOT NULL
                                                         OR location_id IS NOT NULL );

ALTER TYPE item_type ADD VALUE IF NOT EXISTS 'location';
