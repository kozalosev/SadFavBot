CREATE TABLE IF NOT EXISTS Links (
    id serial PRIMARY KEY,
    uid bigint NOT NULL REFERENCES Users(uid),
    alias_id int NOT NULL REFERENCES Aliases(id),
    linked_alias_id int NOT NULL REFERENCES Aliases(id),

    CONSTRAINT uk_links_uid_alias UNIQUE (uid, alias_id)
);

CREATE OR REPLACE FUNCTION ensure_link_not_existing_alias()
    RETURNS TRIGGER
    LANGUAGE PLPGSQL
AS $$
BEGIN
    IF exists(SELECT id FROM items WHERE uid = NEW.uid AND alias = NEW.alias_id) THEN
        RAISE EXCEPTION 'Insertion of the link with the same name as an already existing alias is forbidden';
    END IF;
    RETURN NEW;
END;
$$;

CREATE OR REPLACE TRIGGER trg_ensure_link_not_existing_alias BEFORE INSERT OR UPDATE ON Links
    FOR EACH ROW EXECUTE FUNCTION ensure_link_not_existing_alias()
