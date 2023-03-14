CREATE TABLE IF NOT EXISTS Aliases (
    id serial PRIMARY KEY,
    name varchar(128) NOT NULL UNIQUE CHECK ( name !~ '[•@|\n]' )
);
