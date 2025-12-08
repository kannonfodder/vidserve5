# Feed Service Implementation

This implementation adds a personal feed service to vidserve5 that allows users to create subscriptions to Redgifs searches and creators, with automatic background updates.

## Setup Instructions

### 1. Run Database Migrations

The feed service requires PostgreSQL. Execute the migrations manually:

```bash
psql whutbot < api-go/db/migrations.sql
```

This creates the following tables:

- `users` - User accounts with UUID primary keys
- `feed_subscriptions` - User's saved searches/creators with cursor tracking
- `feed_items` - Cached video results from subscriptions

### 2. Configure Environment Variables

Add the following to your `.env` file:

```bash
DATABASE_URL=postgres://username:password@localhost:5432/whutbot
```

### 3. Build and Run

```bash
cd api-go
go build
./api-go
```

The background worker will start automatically and run every 10 minutes.

## Architecture

### Database Schema

**users**

- `id` (UUID) - Primary key
- `username` (TEXT) - Username
- `created_at` (TIMESTAMP) - Account creation time

**feed_subscriptions**

- `id` (UUID) - Primary key
- `user_id` (UUID) - Foreign key to users
- `type` (TEXT) - "tag" or "creator"
- `search_term` (TEXT) - Search query or creator username
- `last_video_id` (TEXT) - Video ID cursor for deduplication
- `is_initialized` (BOOLEAN) - Whether initial backfill completed
- `created_at` (TIMESTAMP) - Subscription creation time

**feed_items**

- `id` (UUID) - Primary key
- `subscription_id` (UUID) - Foreign key to feed_subscriptions
- `user_id` (UUID) - Foreign key to users (for fast user feed queries)
- `video_id` (TEXT) - Redgifs video ID
- `url` (TEXT) - Video URL
- `username` (TEXT) - Creator username
- `timestamp` (TIMESTAMP) - Video timestamp
- `created_at` (TIMESTAMP) - When item was added to feed
- UNIQUE constraint on `(subscription_id, video_id)` - Prevents duplicates per subscription

### Code Structure

**db/db.go**

- `InitDB()` - Initializes PostgreSQL connection pool using pgxpool
- Requires `DATABASE_URL` environment variable

**auth/user.go**

- Updated `User` struct with `Id` field (UUID)
- `generateUserID()` - Creates UUID for new users
- User ID is stored in encrypted cookie

**feedsvc/subscriptions.go**

- `CreateSubscription()` - Adds new search/creator to user's feed
- `DeleteSubscription()` - Removes subscription
- `ListUserSubscriptions()` - Gets user's active subscriptions
- `GetAllSubscriptions()` - Gets all subscriptions for worker processing

**feedsvc/fetcher.go**

- `FetchAndStore()` - Main function to fetch videos and store in DB
- `fetchInitialVideos()` - Gets first 20 videos for new subscriptions
- `fetchNewVideos()` - Gets videos since last check using cursor
- Video ID comparison for deduplication

**feedsvc/worker.go**

- `StartWorker()` - Ticker-based worker running every 10 minutes
- `runWorkerCycle()` - Processes all subscriptions sequentially
- `runRetentionCleanup()` - Enforces retention policy (50 items minimum, 30 days max age)
- Continues on errors with logging

**feedsvc/feed.go**

- `GetUserFeed()` - Retrieves paginated feed items for display
- `GetUserFeedCount()` - Gets total item count for pagination

## Key Features

### Cursor-Based Deduplication

- Uses `last_video_id` to track newest video per subscription
- When fetching, iterates pages until finding `last_video_id`
- Only stores videos newer than cursor
- Efficient for subscriptions with frequent updates

### Initial Backfill

- New subscriptions fetch 20 most recent videos
- `is_initialized` flag tracks backfill status
- Subsequent fetches only get new content

### Retention Policy

- Guarantees minimum 50 items per user's entire feed
- Deletes items older than 30 days (only if user has 50+ items)
- Runs after each worker cycle
- Per-user cleanup ensures fair distribution

### Error Handling

- Worker continues processing if one subscription fails
- All errors logged with context
- Database connection failures are fatal at startup
- API failures logged but don't halt worker

### Background Worker

- Runs immediately on startup for quick initial population
- 10-minute interval configurable in `worker.go`
- Single Redgifs client reused across subscriptions
- Sequential processing (simple, predictable)

## Usage Examples

### Testing the Service Manually

You can test the feed service by inserting a user and subscriptions directly:

```sql
-- Create a test user
INSERT INTO users (id, username)
VALUES ('550e8400-e29b-41d4-a716-446655440000', 'testuser');

-- Create a tag subscription
INSERT INTO feed_subscriptions (id, user_id, type, search_term, is_initialized)
VALUES (
    '660e8400-e29b-41d4-a716-446655440000',
    '550e8400-e29b-41d4-a716-446655440000',
    'tag',
    'Gay|Twink',
    false
);

-- Create a creator subscription
INSERT INTO feed_subscriptions (id, user_id, type, search_term, is_initialized)
VALUES (
    '770e8400-e29b-41d4-a716-446655440000',
    '550e8400-e29b-41d4-a716-446655440000',
    'creator',
    'bbc21344',
    false
);

-- Wait for the worker to run (every 10 minutes) or restart the app
-- Then check the feed items:
SELECT video_id, username, timestamp
FROM feed_items
WHERE user_id = '550e8400-e29b-41d4-a716-446655440000'
ORDER BY timestamp DESC
LIMIT 20;
```

### Create a Tag Subscription

```go
import "kannonfoundry/api-go/feedsvc"

sub, err := feedsvc.CreateSubscription(
    dbPool,
    userID,
    "tag",
    "Strap On|Gay|Twink",
)
```

### Create a Creator Subscription

```go
sub, err := feedsvc.CreateSubscription(
    dbPool,
    userID,
    "creator",
    "bbc21344",
)
```

### Get User's Feed

```go
items, err := feedsvc.GetUserFeed(dbPool, userID, 20, 0) // limit=20, offset=0
for _, item := range items {
    fmt.Printf("Video: %s by %s\n", item.VideoId, item.Username)
}
```

### List User's Subscriptions

```go
subs, err := feedsvc.ListUserSubscriptions(dbPool, userID)
for _, sub := range subs {
    fmt.Printf("Subscription: %s - %s\n", sub.Type, sub.SearchTerm)
}
```

## Future Enhancements

- [ ] Add HTTP API endpoints for subscription management
- [ ] Implement HTMX-based UI for feed viewing
- [ ] Add WebSocket notifications for new feed items
- [ ] Parallel subscription processing with rate limiting
- [ ] Per-subscription item limits (not just global)
- [ ] Feed filtering by subscription type
- [ ] Export feed to RSS
- [ ] Support for multiple API providers (Rule34, etc.)
