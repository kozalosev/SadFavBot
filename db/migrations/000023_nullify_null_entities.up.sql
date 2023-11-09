UPDATE Texts SET entities = 'null' WHERE entities IS NULL;
ALTER TABLE Texts ALTER COLUMN entities SET NOT NULL;
