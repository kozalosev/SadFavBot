CREATE TABLE IF NOT EXISTS Texts (
    id serial PRIMARY KEY,
    text varchar(10000) NOT NULL UNIQUE
)
