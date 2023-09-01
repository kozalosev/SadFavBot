ALTER TABLE Favs DROP CONSTRAINT IF EXISTS fk_favs_uid;
ALTER TABLE Links DROP CONSTRAINT IF EXISTS fk_links_uid;
ALTER TABLE Packages DROP CONSTRAINT IF EXISTS fk_packages_owner_uid;
ALTER TABLE Alias_visibility DROP CONSTRAINT IF EXISTS fk_alias_visibility_uid;
