package repository

import (
	"context"
	"database/sql"

	"Mmessenger/internal/models"
)

type MessageRepository struct {
	db *sql.DB
}

func NewMessageRepository(db *sql.DB) *MessageRepository {
	return &MessageRepository{db: db}
}

func (r *MessageRepository) Create(ctx context.Context, msg *models.Message) error {
	query := `
		INSERT INTO messages (room_id, sender_id, content, message_type, file_url, thumbnail_url)
		VALUES (?, ?, ?, ?, ?, ?)
	`
	result, err := r.db.ExecContext(ctx, query,
		msg.RoomID, msg.SenderID, msg.Content, msg.MessageType, msg.FileURL, msg.ThumbnailURL,
	)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	msg.ID = uint64(id)

	// Fetch created_at from database
	err = r.db.QueryRowContext(ctx, "SELECT created_at FROM messages WHERE id = ?", msg.ID).Scan(&msg.CreatedAt)
	return err
}

func (r *MessageRepository) GetByID(ctx context.Context, id uint64) (*models.Message, error) {
	query := `
		SELECT id, room_id, sender_id, content, message_type, file_url, thumbnail_url, is_edited, is_deleted, created_at, updated_at
		FROM messages WHERE id = ?
	`
	msg := &models.Message{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&msg.ID, &msg.RoomID, &msg.SenderID, &msg.Content, &msg.MessageType,
		&msg.FileURL, &msg.ThumbnailURL, &msg.IsEdited, &msg.IsDeleted,
		&msg.CreatedAt, &msg.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return msg, nil
}

func (r *MessageRepository) GetByRoomID(ctx context.Context, roomID uint64, limit, offset int) ([]*models.Message, error) {
	query := `
		SELECT id, room_id, sender_id, content, message_type, file_url, thumbnail_url, is_edited, is_deleted, created_at, updated_at
		FROM messages
		WHERE room_id = ? AND is_deleted = FALSE
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`
	rows, err := r.db.QueryContext(ctx, query, roomID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []*models.Message
	for rows.Next() {
		msg := &models.Message{}
		err := rows.Scan(
			&msg.ID, &msg.RoomID, &msg.SenderID, &msg.Content, &msg.MessageType,
			&msg.FileURL, &msg.ThumbnailURL, &msg.IsEdited, &msg.IsDeleted,
			&msg.CreatedAt, &msg.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		messages = append(messages, msg)
	}

	// Reverse to get chronological order
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}
	return messages, nil
}

// GetByRoomIDAfter returns messages after the given message ID (for fetching missed messages)
func (r *MessageRepository) GetByRoomIDAfter(ctx context.Context, roomID uint64, afterID uint64, limit int) ([]*models.Message, error) {
	query := `
		SELECT id, room_id, sender_id, content, message_type, file_url, thumbnail_url, is_edited, is_deleted, created_at, updated_at
		FROM messages
		WHERE room_id = ? AND id > ? AND is_deleted = FALSE
		ORDER BY id ASC
		LIMIT ?
	`
	rows, err := r.db.QueryContext(ctx, query, roomID, afterID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []*models.Message
	for rows.Next() {
		msg := &models.Message{}
		err := rows.Scan(
			&msg.ID, &msg.RoomID, &msg.SenderID, &msg.Content, &msg.MessageType,
			&msg.FileURL, &msg.ThumbnailURL, &msg.IsEdited, &msg.IsDeleted,
			&msg.CreatedAt, &msg.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		messages = append(messages, msg)
	}

	return messages, nil
}

func (r *MessageRepository) Update(ctx context.Context, msg *models.Message) error {
	query := `
		UPDATE messages SET content = ?, is_edited = TRUE, updated_at = NOW()
		WHERE id = ?
	`
	_, err := r.db.ExecContext(ctx, query, msg.Content, msg.ID)
	return err
}

func (r *MessageRepository) Delete(ctx context.Context, id uint64) error {
	query := `UPDATE messages SET is_deleted = TRUE, updated_at = NOW() WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

// GetUnreadCount returns the number of room members who haven't read the message yet
func (r *MessageRepository) GetUnreadCount(ctx context.Context, roomID uint64, messageCreatedAt interface{}, senderID uint64) (int, error) {
	query := `
		SELECT COUNT(*) FROM room_members
		WHERE room_id = ?
		AND user_id != ?
		AND (last_read_at IS NULL OR last_read_at < ?)
	`
	var count int
	err := r.db.QueryRowContext(ctx, query, roomID, senderID, messageCreatedAt).Scan(&count)
	return count, err
}

// GetUnreadCountForUser returns the number of unread messages for a user in a specific room
func (r *MessageRepository) GetUnreadCountForUser(ctx context.Context, roomID uint64, userID uint64) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM messages m
		JOIN room_members rm ON rm.room_id = m.room_id AND rm.user_id = ?
		WHERE m.room_id = ?
		AND m.is_deleted = FALSE
		AND m.sender_id != ?
		AND (rm.last_read_at IS NULL OR m.created_at > rm.last_read_at)
	`
	var count int
	err := r.db.QueryRowContext(ctx, query, userID, roomID, userID).Scan(&count)
	return count, err
}

// GetUnreadCountsForUser returns the unread message counts for all rooms of a user
func (r *MessageRepository) GetUnreadCountsForUser(ctx context.Context, userID uint64) (map[uint64]int, error) {
	query := `
		SELECT rm.room_id, COUNT(m.id) as unread_count
		FROM room_members rm
		LEFT JOIN messages m ON m.room_id = rm.room_id
			AND m.is_deleted = FALSE
			AND m.sender_id != ?
			AND (rm.last_read_at IS NULL OR m.created_at > rm.last_read_at)
		WHERE rm.user_id = ?
		GROUP BY rm.room_id
	`
	rows, err := r.db.QueryContext(ctx, query, userID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	counts := make(map[uint64]int)
	for rows.Next() {
		var roomID uint64
		var count int
		if err := rows.Scan(&roomID, &count); err != nil {
			return nil, err
		}
		counts[roomID] = count
	}
	return counts, nil
}
