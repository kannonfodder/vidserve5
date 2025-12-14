package feedsvc

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"kannonfoundry/api-go/api"
	"kannonfoundry/api-go/api/redgifs"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type VideoItem struct {
	Id        string
	VideoId   string
	Url       string
	Username  string
	Timestamp time.Time
}

// FetchAndStore fetches new videos for a subscription and stores them in the database
func FetchAndStore(db *pgxpool.Pool, subscription *Subscription, rgClient *redgifs.RedGifsClient) error {
	var videos []VideoItem
	var err error

	// Determine fetch strategy based on initialization status
	if !subscription.IsInitialized {
		// First time fetch: get initial 20 items
		videos, err = fetchInitialVideos(subscription, rgClient)
		if err != nil {
			return fmt.Errorf("failed to fetch initial videos: %w", err)
		}

		// Mark as initialized and set last_video_id to the newest video
		if len(videos) > 0 {
			if err := markSubscriptionInitialized(db, subscription.Id, videos[0].VideoId); err != nil {
				return fmt.Errorf("failed to mark subscription as initialized: %w", err)
			}
			subscription.IsInitialized = true
			subscription.LastVideoId = &videos[0].VideoId
		}
	} else {
		// Regular fetch: get new videos since last check
		videos, err = fetchNewVideos(subscription, rgClient)
		if err != nil {
			return fmt.Errorf("failed to fetch new videos: %w", err)
		}

		// Update last_video_id if we got new videos
		if len(videos) > 0 {
			if err := updateLastVideoId(db, subscription.Id, videos[0].VideoId); err != nil {
				return fmt.Errorf("failed to update last video id: %w", err)
			}
			subscription.LastVideoId = &videos[0].VideoId
		}
	}

	// Store videos in database
	if len(videos) > 0 {
		if err := storeVideos(db, subscription, videos); err != nil {
			return fmt.Errorf("failed to store videos: %w", err)
		}
		log.Printf("Stored %d new videos for subscription %s (%s: %s)", len(videos), subscription.Id, subscription.Type, subscription.SearchTerm)
	}

	return nil
}

// fetchInitialVideos gets the first 20 videos for a new subscription
func fetchInitialVideos(subscription *Subscription, rgClient *redgifs.RedGifsClient) ([]VideoItem, error) {
	if subscription.Type == "creator" {
		// Search by user
		files, err := rgClient.SearchByUser(subscription.SearchTerm, 20, 1)
		if err != nil {
			return nil, err
		}
		return apiFilesToVideoItems(files), nil
	} else {
		// Tag search
		tags := strings.Split(subscription.SearchTerm, "|")
		files, err := rgClient.Search(tags, 20, 1)
		if err != nil {
			return nil, err
		}
		return apiFilesToVideoItems(files), nil
	}
}

// fetchNewVideos gets videos since the last check, stopping when we find last_video_id
func fetchNewVideos(subscription *Subscription, rgClient *redgifs.RedGifsClient) ([]VideoItem, error) {
	var allVideos []VideoItem
	page := 1
	maxPages := 5 // Safety limit to prevent infinite loops

	for page <= maxPages {
		var files []api.FileToSend
		var err error

		if subscription.Type == "creator" {
			files, err = rgClient.SearchByUser(subscription.SearchTerm, 20, page)
			if err != nil {
				return nil, err
			}
		} else {
			tags := strings.Split(subscription.SearchTerm, "|")
			files, err = rgClient.Search(tags, 20, page)
			if err != nil {
				return nil, err
			}
		}

		if len(files) == 0 {
			break // No more results
		}

		// Process files and check for last_video_id
		for _, file := range files {
			videoId := file.Name

			// Stop if we've reached the last seen video
			if subscription.LastVideoId != nil && videoId == *subscription.LastVideoId {
				return allVideos, nil
			}

			allVideos = append(allVideos, VideoItem{
				VideoId:   videoId,
				Url:       file.URL,
				Username:  file.Username,
				Timestamp: time.Unix(file.CreatedAt, 0),
			})
		}

		page++
	}

	return allVideos, nil
}

// apiFilesToVideoItems converts API FileToSend structs to VideoItem structs
func apiFilesToVideoItems(files []api.FileToSend) []VideoItem {
	var videos []VideoItem
	for _, file := range files {
		videos = append(videos, VideoItem{
			VideoId:   file.Name,
			Url:       file.URL,
			Username:  file.Username,
			Timestamp: time.Unix(file.CreatedAt, 0),
		})
	}
	return videos
}

// storeVideos inserts videos into the database
func storeVideos(db *pgxpool.Pool, subscription *Subscription, videos []VideoItem) error {
	query := `
		INSERT INTO feed_items (id, subscription_id, user_id, video_id, url, username, timestamp)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (subscription_id, video_id) DO NOTHING
	`

	for _, video := range videos {
		id := uuid.New().String()
		_, err := db.Exec(context.Background(), query,
			id, subscription.Id, subscription.UserId, video.VideoId,
			video.Url, video.Username, video.Timestamp,
		)
		if err != nil {
			return fmt.Errorf("failed to insert video %s: %w", video.VideoId, err)
		}
	}

	return nil
}

// markSubscriptionInitialized marks a subscription as initialized
func markSubscriptionInitialized(db *pgxpool.Pool, subscriptionId, lastVideoId string) error {
	query := `UPDATE feed_subscriptions SET is_initialized = true, last_video_id = $1 WHERE id = $2`
	_, err := db.Exec(context.Background(), query, lastVideoId, subscriptionId)
	return err
}

// updateLastVideoId updates the last_video_id for a subscription
func updateLastVideoId(db *pgxpool.Pool, subscriptionId, lastVideoId string) error {
	query := `UPDATE feed_subscriptions SET last_video_id = $1 WHERE id = $2`
	_, err := db.Exec(context.Background(), query, lastVideoId, subscriptionId)
	return err
}
