-- Add sticker message type to messages table
ALTER TABLE messages
MODIFY COLUMN message_type ENUM('text', 'image', 'file', 'system', 'sticker') DEFAULT 'text';
