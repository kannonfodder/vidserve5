# Plan: Enhance Creator Pages with Feed Subscription

## Current State

- **Creator page** (`components/layout/creator.templ`) only displays username and video grid
- **CreatorResponse** struct captures: `Username`, `Description`, `ProfileImageUrl`
- **Feed system** exists with subscription support for "tag" and "creator" types
- No UI for adding creators to feeds
- No route for feed display

## Required Changes

### 1. **Enhance Creator Page UI** (`components/layout/creator.templ`)

- Display creator profile image
- Display creator description/bio
- Add "Subscribe to Feed" button (HTMX-enabled)
- Show subscription status if user is logged in and already subscribed
- Improve layout with better styling for creator info section

### 2. **Create Subscription Endpoints** (`routes/creators/serve.go` or new file)

- **POST `/creators/{username}/subscribe`** - Create feed subscription for creator
  - Validate user is logged in
  - Call `feedsvc.CreateSubscription()` with type="creator"
  - Return HTMX response to update button state
- **DELETE `/creators/{username}/subscribe`** - Remove feed subscription
  - Delete subscription for user+creator
  - Return HTMX response to update button state
- **GET `/creators/{username}/subscription-status`** - Check if subscribed
  - Query database for existing subscription
  - Return button state for HTMX swap

### 3. **Update Creator Serve Handler** (`routes/creators/serve.go`)

- Pass database pool to route handler
- Check user authentication status
- Query if user has existing subscription to this creator
- Pass subscription status to template

### 4. **Create/Update Feed Display Route** (`routes/feed/serve.go`)

- Implement full feed page displaying user's feed items
- Show paginated video grid from `feedsvc.GetUserFeed()`
- Include infinite scroll or "Load More" with HTMX
- Show which subscription each video came from
- Require authentication

### 5. **Update Main Router** (`main.go`)

- Pass database pool to creator routes
- Register new subscription endpoints
- Update feed route to actual implementation
- Ensure all routes have proper authentication middleware

### 6. **Database Considerations**

- Verify subscription uniqueness constraint (user_id + type + search_term)
- May need to add check to prevent duplicate creator subscriptions

### 7. **UI Components to Create**

- Subscribe button component with loading/subscribed states
- Creator header/profile component
- Feed item card showing video + subscription source
- List of user's subscriptions (for feed page sidebar)

## File Changes Summary

- **Edit**: `api-go/components/layout/creator.templ` - Enhanced creator UI
- **Edit**: `api-go/routes/creators/serve.go` - Add DB pool, subscription status check
- **Create**: `api-go/routes/creators/subscribe.go` - Subscription endpoints
- **Edit**: `api-go/routes/feed/serve.go` - Implement feed display
- **Edit**: `api-go/main.go` - Wire up new routes and DB pool
- **Optional**: Add indexes to ensure subscription uniqueness

## Implementation Order

1. Update creator page template with new data and UI
2. Create subscription HTTP endpoints
3. Update creator route handler with DB and auth
4. Wire routes in main.go
5. Implement feed display page
6. Test subscription flow end-to-end
