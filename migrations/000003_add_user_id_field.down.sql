DROP INDEX CONCURRENTLY idx_user_id ON shorted_links(user_id);
ALTER TABLE shorted_links DROP COLUMN user_id;
-- ALTER TABLE shorted_links DROP COLUMN shorted_url;