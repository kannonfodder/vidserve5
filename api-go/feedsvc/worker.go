package feedsvc

import (
	"context"
	"log"
	"time"

	"kannonfoundry/api-go/api/redgifs"

	"github.com/jackc/pgx/v5/pgxpool"
)

// StartWorker starts the background worker that periodically fetches new videos
func StartWorker(db *pgxpool.Pool) {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()

	log.Println("Feed worker started, will run every 10 minutes")

	// Run immediately on startup
	runWorkerCycle(db)

	// Then run on ticker
	for range ticker.C {
		runWorkerCycle(db)
	}
}

// runWorkerCycle executes one complete worker cycle
func runWorkerCycle(db *pgxpool.Pool) {
	log.Println("Starting feed worker cycle...")

	// Create Redgifs client
	rgClient := redgifs.NewClient()

	// Get all subscriptions
	subscriptions, err := GetAllSubscriptions(db)
	if err != nil {
		log.Printf("ERROR: Failed to get subscriptions: %v", err)
		return
	}

	log.Printf("Processing %d subscriptions", len(subscriptions))

	// Process each subscription
	successCount := 0
	errorCount := 0
	for _, sub := range subscriptions {
		err := FetchAndStore(db, &sub, rgClient)
		if err != nil {
			log.Printf("ERROR: Failed to fetch and store for subscription %s (%s: %s): %v",
				sub.Id, sub.Type, sub.SearchTerm, err)
			errorCount++
		} else {
			successCount++
		}
	}

	log.Printf("Worker cycle completed: %d succeeded, %d failed", successCount, errorCount)

	// Run retention cleanup for all users
	if err := runRetentionCleanup(db); err != nil {
		log.Printf("ERROR: Failed to run retention cleanup: %v", err)
	}
}

// runRetentionCleanup removes old feed items while maintaining minimum 50 items per user
func runRetentionCleanup(db *pgxpool.Pool) error {
	// Get all unique user IDs
	userQuery := `SELECT DISTINCT user_id FROM feed_subscriptions`
	rows, err := db.Query(context.Background(), userQuery)
	if err != nil {
		return err
	}
	defer rows.Close()

	var userIds []string
	for rows.Next() {
		var userId string
		if err := rows.Scan(&userId); err != nil {
			return err
		}
		userIds = append(userIds, userId)
	}

	// Run cleanup for each user
	deleteQuery := `
		DELETE FROM feed_items 
		WHERE user_id = $1 
		AND id NOT IN (
			SELECT id FROM feed_items 
			WHERE user_id = $1 
			ORDER BY timestamp DESC 
			LIMIT 50
		) 
		AND created_at < NOW() - INTERVAL '30 days'
	`

	totalDeleted := 0
	for _, userId := range userIds {
		result, err := db.Exec(context.Background(), deleteQuery, userId)
		if err != nil {
			log.Printf("ERROR: Failed to cleanup for user %s: %v", userId, err)
			continue
		}

		deleted := result.RowsAffected()
		if deleted > 0 {
			log.Printf("Cleaned up %d old items for user %s", deleted, userId)
			totalDeleted += int(deleted)
		}
	}

	if totalDeleted > 0 {
		log.Printf("Retention cleanup completed: removed %d total items", totalDeleted)
	}

	return nil
}
