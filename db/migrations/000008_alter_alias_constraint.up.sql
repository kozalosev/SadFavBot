ALTER TABLE Aliases DROP CONSTRAINT IF EXISTS aliases_name_check;
ALTER TABLE Aliases ADD CONSTRAINT aliases_name_check CHECK ( name !~ '[•@|\n{}[\]:]' );
