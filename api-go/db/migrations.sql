-- Feed Service Database Schema
-- Execute manually with: psql whutbot < db/migrations.sql

-- Users table
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY,
    username TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);

-- Feed subscriptions (user's saved searches and creators)
CREATE TABLE IF NOT EXISTS feed_subscriptions (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    type TEXT NOT NULL CHECK (type IN ('tag', 'creator')),
    search_term TEXT NOT NULL,
    last_video_id TEXT,
    is_initialized BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_feed_subscriptions_user_id ON feed_subscriptions(user_id);
CREATE INDEX IF NOT EXISTS idx_feed_subscriptions_is_initialized ON feed_subscriptions(is_initialized);

-- Feed items (cached videos from subscriptions)
CREATE TABLE IF NOT EXISTS feed_items (
    id UUID PRIMARY KEY,
    subscription_id UUID NOT NULL REFERENCES feed_subscriptions(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    video_id TEXT NOT NULL,
    url TEXT NOT NULL,
    username TEXT NOT NULL,
    timestamp TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE(subscription_id, video_id)
);

CREATE INDEX IF NOT EXISTS idx_feed_items_user_timestamp ON feed_items(user_id, timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_feed_items_subscription ON feed_items(subscription_id);
CREATE INDEX IF NOT EXISTS idx_feed_items_created_at ON feed_items(created_at);
