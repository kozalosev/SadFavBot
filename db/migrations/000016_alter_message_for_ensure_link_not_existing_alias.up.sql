CREATE OR REPLACE FUNCTION ensure_link_not_existing_alias()
    RETURNS TRIGGER
    LANGUAGE PLPGSQL
AS $$
BEGIN
    IF exists(SELECT id FROM items WHERE uid = NEW.uid AND alias = NEW.alias_id) THEN
        RAISE EXCEPTION 'Insertion of the link with the same name as an already existing fav is forbidden';
    END IF;
    RETURN NEW;
END;
$$;
