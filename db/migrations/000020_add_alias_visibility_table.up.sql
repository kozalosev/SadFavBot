CREATE TABLE IF NOT EXISTS Alias_Visibility (
    uid bigint,
    alias_id int,
    hidden bool NOT NULL,

    PRIMARY KEY (uid, alias_id)
);

INSERT INTO Alias_Visibility (uid, alias_id, hidden)
    SELECT DISTINCT uid, alias_id, hidden FROM Favs WHERE hidden = true;

ALTER TABLE Favs DROP COLUMN hidden;
