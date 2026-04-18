ALTER TABLE tasks DROP CONSTRAINT tasks_parent_id_fkey;
ALTER TABLE tasks ADD CONSTRAINT tasks_parent_id_fkey
    FOREIGN KEY (parent_id) REFERENCES tasks(id) ON DELETE CASCADE;