DROP TABLE IF EXISTS Locations;
ALTER TABLE Favs DROP COLUMN IF EXISTS location_id;

ALTER TABLE Favs DROP CONSTRAINT IF EXISTS ck_favs_content;
ALTER TABLE Favs ADD CONSTRAINT ck_favs_content CHECK ( file_id IS NOT NULL AND file_unique_id IS NOT NULL
    OR text_id IS NOT NULL );
