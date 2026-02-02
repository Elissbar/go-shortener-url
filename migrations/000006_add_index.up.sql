CREATE INDEX IF NOT EXISTS idx_user_id ON shorted_links(user_id);
CREATE INDEX IF NOT EXISTS idx_url ON shorted_links(url);