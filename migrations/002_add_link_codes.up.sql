CREATE TABLE IF NOT EXISTS link_codes (
    id SERIAL PRIMARY KEY,
    user_id INTEGER UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    code VARCHAR(32) UNIQUE NOT NULL,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_link_codes_code ON link_codes(code);
CREATE INDEX IF NOT EXISTS idx_link_codes_user_id ON link_codes(user_id);
CREATE INDEX IF NOT EXISTS idx_link_codes_expires_at ON link_codes(expires_at);
