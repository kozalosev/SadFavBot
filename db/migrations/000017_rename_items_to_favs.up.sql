ALTER TABLE Items RENAME TO Favs;
ALTER TABLE Favs RENAME COLUMN alias TO alias_id;
ALTER TABLE Favs RENAME COLUMN text TO text_id;

ALTER TABLE Favs RENAME CONSTRAINT fk_item_alias TO fk_favs_alias;
ALTER TABLE Favs RENAME CONSTRAINT fk_item_text TO fk_favs_text;
ALTER TABLE Favs RENAME CONSTRAINT ck_item_data TO ck_favs_content;

ALTER INDEX items_pkey RENAME TO favs_pkey;
ALTER INDEX uk_item_file RENAME TO uk_favs_file;
ALTER INDEX uk_item_text RENAME TO uk_favs_text;
ALTER INDEX idx_items_uid_alias RENAME TO idx_favs_uid_alias;

CREATE OR REPLACE FUNCTION ensure_link_not_existing_alias()
    RETURNS TRIGGER
    LANGUAGE PLPGSQL
AS $$
BEGIN
    IF exists(SELECT id FROM favs WHERE uid = NEW.uid AND alias_id = NEW.alias_id) THEN
        RAISE EXCEPTION 'Insertion of the link with the same name as an already existing fav is forbidden';
    END IF;
    RETURN NEW;
END;
$$;
