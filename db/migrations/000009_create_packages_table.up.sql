CREATE TABLE IF NOT EXISTS Packages (
    id serial PRIMARY KEY,
    owner_uid bigint NOT NULL,
    name varchar(255) NOT NULL CHECK ( name !~ '[•@|\n{}[\] ]' ),

    FOREIGN KEY (owner_uid) REFERENCES Users(uid),
    CONSTRAINT uix_package_owner_name UNIQUE (owner_uid, name)
);

CREATE TABLE IF NOT EXISTS Package_Aliases (
    package_id integer NOT NULL,
    alias_id integer NOT NULL,

    FOREIGN KEY (package_id) REFERENCES Packages(id) ON DELETE CASCADE,
    FOREIGN KEY (alias_id) REFERENCES Aliases(id),
    CONSTRAINT uix_package_alias UNIQUE (package_id, alias_id)
);

CREATE INDEX IF NOT EXISTS idx_package_aliases_package_id ON Package_Aliases(package_id);
