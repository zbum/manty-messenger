-- Add keycloak_id column for SSO integration
ALTER TABLE users ADD COLUMN keycloak_id VARCHAR(255) NULL UNIQUE AFTER id;

-- Make password_hash nullable since SSO users won't have local passwords
ALTER TABLE users MODIFY COLUMN password_hash VARCHAR(255) NULL;
