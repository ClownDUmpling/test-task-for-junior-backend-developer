ALTER TABLE tasks ADD COLUMN recurrence_type TEXT;
ALTER TABLE tasks ADD COLUMN recurrence_every_n_days INT;
ALTER TABLE tasks ADD COLUMN recurrence_month_days INT[];
ALTER TABLE tasks ADD COLUMN recurrence_dates DATE[];
ALTER TABLE tasks ADD COLUMN recurrence_parity TEXT;
ALTER TABLE tasks ADD COLUMN scheduled_date DATE;
ALTER TABLE tasks ADD COLUMN parent_id BIGINT REFERENCES tasks(id);