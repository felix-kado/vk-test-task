-- Add author_login column to ads table for denormalization
ALTER TABLE ads ADD COLUMN author_login VARCHAR(32) NOT NULL DEFAULT '';

-- Update existing ads with author login from users table
UPDATE ads 
SET author_login = users.login 
FROM users 
WHERE ads.user_id = users.id;

-- Remove default constraint after data migration
ALTER TABLE ads ALTER COLUMN author_login DROP DEFAULT;
