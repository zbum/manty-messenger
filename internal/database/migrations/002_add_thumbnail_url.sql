-- Add thumbnail_url column to messages table
ALTER TABLE messages ADD COLUMN thumbnail_url VARCHAR(512) NULL AFTER file_url;
