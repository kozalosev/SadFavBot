ALTER TABLE Favs ADD COLUMN IF NOT EXISTS hidden bool NOT NULL DEFAULT false;

UPDATE Favs SET hidden = AV.hidden
    FROM Favs JOIN Alias_Visibility AV ON Favs.uid = AV.uid AND Favs.alias_id = AV.alias_id;

DROP TABLE Alias_Visibility;
