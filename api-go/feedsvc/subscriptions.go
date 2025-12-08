package feedsvc

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Subscription struct {
	Id            string
	UserId        string
	Type          string // "tag" or "creator"
	SearchTerm    string
	LastVideoId   string
	IsInitialized bool
}

// CreateSubscription adds a new feed subscription for a user
func CreateSubscription(db *pgxpool.Pool, userID, subscriptionType, searchTerm string) (*Subscription, error) {
	id := uuid.New().String()

	query := `
		INSERT INTO feed_subscriptions (id, user_id, type, search_term, is_initialized)
		VALUES ($1, $2, $3, $4, false)
		RETURNING id, user_id, type, search_term, last_video_id, is_initialized
	`

	sub := &Subscription{}
	err := db.QueryRow(context.Background(), query, id, userID, subscriptionType, searchTerm).Scan(
		&sub.Id, &sub.UserId, &sub.Type, &sub.SearchTerm, &sub.LastVideoId, &sub.IsInitialized,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create subscription: %w", err)
	}

	return sub, nil
}

// DeleteSubscription removes a feed subscription
func DeleteSubscription(db *pgxpool.Pool, subscriptionID string) error {
	query := `DELETE FROM feed_subscriptions WHERE id = $1`

	result, err := db.Exec(context.Background(), query, subscriptionID)
	if err != nil {
		return fmt.Errorf("failed to delete subscription: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("subscription not found")
	}

	return nil
}

// ListUserSubscriptions retrieves all subscriptions for a user
func ListUserSubscriptions(db *pgxpool.Pool, userID string) ([]Subscription, error) {
	query := `
		SELECT id, user_id, type, search_term, last_video_id, is_initialized
		FROM feed_subscriptions
		WHERE user_id = $1
		ORDER BY created_at DESC
	`

	rows, err := db.Query(context.Background(), query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list subscriptions: %w", err)
	}
	defer rows.Close()

	var subscriptions []Subscription
	for rows.Next() {
		var sub Subscription
		err := rows.Scan(&sub.Id, &sub.UserId, &sub.Type, &sub.SearchTerm, &sub.LastVideoId, &sub.IsInitialized)
		if err != nil {
			return nil, fmt.Errorf("failed to scan subscription: %w", err)
		}
		subscriptions = append(subscriptions, sub)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating subscriptions: %w", err)
	}

	return subscriptions, nil
}

// GetAllSubscriptions retrieves all subscriptions for the worker to process
func GetAllSubscriptions(db *pgxpool.Pool) ([]Subscription, error) {
	query := `
		SELECT id, user_id, type, search_term, last_video_id, is_initialized
		FROM feed_subscriptions
		ORDER BY created_at ASC
	`

	rows, err := db.Query(context.Background(), query)
	if err != nil {
		return nil, fmt.Errorf("failed to get all subscriptions: %w", err)
	}
	defer rows.Close()

	var subscriptions []Subscription
	for rows.Next() {
		var sub Subscription
		err := rows.Scan(&sub.Id, &sub.UserId, &sub.Type, &sub.SearchTerm, &sub.LastVideoId, &sub.IsInitialized)
		if err != nil {
			return nil, fmt.Errorf("failed to scan subscription: %w", err)
		}
		subscriptions = append(subscriptions, sub)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating subscriptions: %w", err)
	}

	return subscriptions, nil
}
