-- Create database
CREATE DATABASE IF NOT EXISTS manty_messenger
    CHARACTER SET utf8mb4
    COLLATE utf8mb4_unicode_ci;

-- Create user and grant privileges
DROP USER IF EXISTS 'manty'@'%';
CREATE USER 'manty'@'%' IDENTIFIED WITH mysql_native_password BY 'wjdwlqja2@';
GRANT ALL PRIVILEGES ON manty_messenger.* TO 'manty'@'%';
FLUSH PRIVILEGES;

-- Verify
SHOW DATABASES LIKE 'manty_messenger';
SELECT User, Host, plugin FROM mysql.user WHERE User = 'manty';
