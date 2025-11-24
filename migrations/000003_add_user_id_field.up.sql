ALTER TABLE shorted_links ADD COLUMN user_id VARCHAR(36);
-- ALTER TABLE shorted_links ADD COLUMN shorted_url VARCHAR(255);
CREATE INDEX idx_user_id ON shorted_links(user_id);