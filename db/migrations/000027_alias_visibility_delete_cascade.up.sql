ALTER TABLE alias_visibility ADD CONSTRAINT alias_visibility_alias_id_fkey FOREIGN KEY (alias_id) REFERENCES aliases(id) ON DELETE CASCADE;
