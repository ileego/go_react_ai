ALTER TABLE users
    DROP COLUMN IF EXISTS github_id,
    DROP COLUMN IF EXISTS avatar_url;
