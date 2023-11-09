ALTER TABLE Texts ALTER COLUMN entities DROP NOT NULL;
UPDATE Texts SET entities = NULL WHERE entities = 'null';
