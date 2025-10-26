-- ===========================================
-- 🧱 NomNom Hub – Core Database Schema
-- ===========================================

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- -------------------------------
--  users
-- -------------------------------
CREATE TABLE users (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    line_id         TEXT UNIQUE NOT NULL,
    display_name    TEXT,
    created_at      TIMESTAMP WITH TIME ZONE DEFAULT now(),
    updated_at      TIMESTAMP WITH TIME ZONE DEFAULT now()
);

-- -------------------------------
--  places
-- -------------------------------
CREATE TABLE places (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name            TEXT NOT NULL,
    url             TEXT NOT NULL,
    added_by        UUID REFERENCES users(id) ON DELETE SET NULL,
    created_at      TIMESTAMP WITH TIME ZONE DEFAULT now(),
    updated_at      TIMESTAMP WITH TIME ZONE DEFAULT now()
);

CREATE UNIQUE INDEX idx_places_url_unique ON places (url);

-- -------------------------------
--  votes
-- -------------------------------
CREATE TABLE votes (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    place_id        UUID NOT NULL REFERENCES places(id) ON DELETE CASCADE,
    user_id         UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    value           SMALLINT NOT NULL CHECK (value BETWEEN -1 AND 1),
    created_at      TIMESTAMP WITH TIME ZONE DEFAULT now(),
    updated_at      TIMESTAMP WITH TIME ZONE DEFAULT now(),
    UNIQUE (place_id, user_id)
);

-- value mapping:
-- -1 = Dislike
--  0 = Meh
-- +1 = Interested

-- -------------------------------
--  tags
-- -------------------------------
CREATE TABLE tags (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name            TEXT UNIQUE NOT NULL
);

-- -------------------------------
--  place_tags
-- -------------------------------
CREATE TABLE place_tags (
    place_id        UUID NOT NULL REFERENCES places(id) ON DELETE CASCADE,
    tag_id          UUID NOT NULL REFERENCES tags(id) ON DELETE CASCADE,
    PRIMARY KEY (place_id, tag_id)
);

-- -------------------------------
--  triggers for updated_at
-- -------------------------------
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = now();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_users_updated_at
BEFORE UPDATE ON users
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER trg_places_updated_at
BEFORE UPDATE ON places
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER trg_votes_updated_at
BEFORE UPDATE ON votes
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();

-- -------------------------------
--  views
-- -------------------------------
CREATE OR REPLACE VIEW place_summary AS
SELECT
    p.id,
    p.name,
    p.url,
    COALESCE(ROUND(AVG(v.value)::numeric, 2), 0) AS score,
    COUNT(v.id) AS total_votes,
    array_agg(DISTINCT t.name) AS tags
FROM places p
LEFT JOIN votes v ON v.place_id = p.id
LEFT JOIN place_tags pt ON pt.place_id = p.id
LEFT JOIN tags t ON t.id = pt.tag_id
GROUP BY p.id;
