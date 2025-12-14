package feedsvc

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// GetUserFeed retrieves paginated feed items for a user
func GetUserFeed(db *pgxpool.Pool, userID string, limit, offset int) ([]VideoItem, error) {
	query := `
		SELECT id, video_id, url, username, timestamp
		FROM feed_items
		WHERE user_id = $1
		ORDER BY timestamp DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := db.Query(context.Background(), query, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query feed items: %w", err)
	}
	defer rows.Close()

	var items []VideoItem
	for rows.Next() {
		var item VideoItem
		err := rows.Scan(&item.Id, &item.VideoId, &item.Url, &item.Username, &item.Timestamp)
		if err != nil {
			return nil, fmt.Errorf("failed to scan feed item: %w", err)
		}
		items = append(items, item)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating feed items: %w", err)
	}

	return items, nil
}

// GetUserFeedCount returns the total number of feed items for a user
func GetUserFeedCount(db *pgxpool.Pool, userID string) (int, error) {
	query := `SELECT COUNT(*) FROM feed_items WHERE user_id = $1`

	var count int
	err := db.QueryRow(context.Background(), query, userID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count feed items: %w", err)
	}

	return count, nil
}
