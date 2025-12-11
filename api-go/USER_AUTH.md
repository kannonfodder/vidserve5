# User Authentication System

This document describes the updated user authentication system that replaces hardcoded credentials with proper database-backed user accounts.

## Overview

The authentication system now supports:

- ✅ Database-backed user accounts (PostgreSQL)
- ✅ Bcrypt password hashing for security
- ✅ User registration with validation
- ✅ Secure session cookies (HttpOnly, 7-day expiry)
- ✅ Unique username constraints
- ✅ Password strength requirements (minimum 8 characters)

## Database Schema

The `users` table has been updated with password support:

```sql
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY,
    username TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);
```

## Migration Steps

### 1. Update Database Schema

If you already ran the previous migrations, you need to add the `password_hash` column:

```bash
# Connect to your database
psql whutbot

# Add password_hash column and unique constraint
ALTER TABLE users ADD COLUMN IF NOT EXISTS password_hash TEXT;
ALTER TABLE users ADD CONSTRAINT users_username_unique UNIQUE (username);
```

Or drop and recreate (WARNING: This will delete all existing data):

```bash
psql whutbot < api-go/db/migrations.sql
```

### 2. Create Initial User Accounts

You can create users via the registration page at `/register`, or manually via SQL:

```sql
-- Example: Create an admin user with password "securepassword123"
-- First, hash the password using bcrypt (cost 10)
-- You can use an online bcrypt generator or Go code

-- Using Go to generate hash:
-- echo 'package main; import ("fmt"; "golang.org/x/crypto/bcrypt"); func main() { h, _ := bcrypt.GenerateFromPassword([]byte("securepassword123"), 10); fmt.Println(string(h)) }' > hash.go
-- go run hash.go

INSERT INTO users (id, username, password_hash)
VALUES (
    gen_random_uuid(),
    'admin',
    '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy'  -- "password"
);
```

### 3. Restart Application

The application will now use database authentication:

```bash
cd api-go
go build
./api-go
```

## Architecture

### Files Modified/Created

**Database Layer:**

- `db/migrations.sql` - Updated users table schema

**Authentication Layer:**

- `auth/user_service.go` - NEW: Database user operations (CreateUser, AuthenticateUser, GetUserByID, GetUserByUsername)
- `auth/user.go` - Updated: Replaced hardcoded `LoginUser` with `LoginUserWithDB`

**Routes:**

- `routes/login/serve.go` - Updated: Uses database authentication
- `routes/register/serve.go` - NEW: User registration handling

**Components:**

- `components/register.templ` - NEW: Registration form template
- `components/login.templ` - Updated: Added link to registration

**Main Application:**

- `main.go` - Updated: Pass database pool to login and register routes

### Authentication Flow

#### Registration Flow:

1. User visits `/register`
2. Fills out username, password, confirm password
3. System validates:
   - All fields required
   - Password minimum 8 characters
   - Passwords match
4. Password hashed with bcrypt (cost 10)
5. User created in database with UUID
6. User auto-logged in with session cookie
7. Redirect to home page

#### Login Flow:

1. User visits `/login`
2. Enters username and password
3. System queries database for username
4. Compares submitted password with stored bcrypt hash
5. On success: Creates encrypted session cookie with user data
6. Redirect to home page

#### Session Validation:

1. User makes request with session cookie
2. `auth.IsLoggedIn(r)` decrypts cookie
3. Returns User object with Id and Username
4. No database lookup needed for each request (cookie contains user data)

## Security Features

### Password Hashing

- Uses bcrypt with default cost (10)
- Salt automatically generated per password
- Resistant to rainbow table and brute force attacks

### Session Cookies

- `HttpOnly` flag prevents JavaScript access
- 7-day expiration (`MaxAge: 86400 * 7`)
- Path set to `/` for site-wide access
- Encrypted with AES-256-GCM
- Contains user ID and username (no sensitive data)

### Database

- UNIQUE constraint on username prevents duplicates
- Password hash never exposed in API responses
- UUID primary keys prevent enumeration attacks

## API Reference

### Authentication Functions

**`auth.CreateUser(db, username, password)`**

```go
user, err := auth.CreateUser(dbPool, "johndoe", "mypassword123")
// Returns: *User, error
```

**`auth.AuthenticateUser(db, username, password)`**

```go
user, err := auth.AuthenticateUser(dbPool, "johndoe", "mypassword123")
// Returns: *User, error (nil user if invalid credentials)
```

**`auth.LoginUserWithDB(db, username, password)`**

```go
cookie, err := auth.LoginUserWithDB(dbPool, "johndoe", "mypassword123")
// Returns: encrypted cookie string, error
```

**`auth.GetUserByID(db, userID)`**

```go
user, err := auth.GetUserByID(dbPool, "550e8400-e29b-41d4-a716-446655440000")
// Returns: *User, error
```

**`auth.GetUserByUsername(db, username)`**

```go
user, err := auth.GetUserByUsername(dbPool, "johndoe")
// Returns: *User, error
```

**`auth.IsLoggedIn(r)`**

```go
user := auth.IsLoggedIn(r) // r is *http.Request
// Returns: User (check user.IsEmpty() to see if logged in)
```

## Usage Examples

### Protecting Routes

```go
func MyProtectedRoute(w http.ResponseWriter, r *http.Request) {
    user := auth.IsLoggedIn(r)
    if user.IsEmpty() {
        http.Redirect(w, r, "/login", http.StatusSeeOther)
        return
    }

    // User is authenticated
    fmt.Fprintf(w, "Hello, %s (ID: %s)", user.Username, user.Id)
}
```

### Manual User Creation

```go
func CreateAdminUser(db *pgxpool.Pool) {
    user, err := auth.CreateUser(db, "admin", "strongpassword123")
    if err != nil {
        log.Fatalf("Failed to create admin: %v", err)
    }
    log.Printf("Created admin user: %s", user.Id)
}
```

### Custom Login Handler

```go
func CustomLogin(w http.ResponseWriter, r *http.Request) {
    username := r.FormValue("username")
    password := r.FormValue("password")

    cookie, err := auth.LoginUserWithDB(dbPool, username, password)
    if err != nil {
        http.Error(w, "Invalid credentials", http.StatusUnauthorized)
        return
    }

    http.SetCookie(w, &http.Cookie{
        Name:     "session_id",
        Value:    cookie,
        Path:     "/",
        HttpOnly: true,
        MaxAge:   86400 * 7,
    })

    w.Write([]byte("Login successful"))
}
```

## Routes

- `GET /register` - Registration form
- `POST /register` - Process registration
- `GET /login` - Login form
- `POST /login` - Process login
- `POST /logout` - Clear session and logout

## Environment Variables

Ensure `SECRET_KEY` is set for cookie encryption:

```bash
# .env or .key.env
SECRET_KEY=your-32-character-secret-key-here-change-this
DATABASE_URL=postgres://username:password@localhost:5432/whutbot
```

## Testing

### Create a Test User

```bash
# Start the application
./api-go

# In browser, visit:
http://localhost:8080/register

# Or use curl:
curl -X POST http://localhost:8080/register \
  -d "username=testuser" \
  -d "password=testpass123" \
  -d "confirm_password=testpass123"
```

### Verify User in Database

```bash
psql whutbot -c "SELECT id, username, created_at FROM users;"
```

### Test Login

```bash
curl -X POST http://localhost:8080/login \
  -d "username=testuser" \
  -d "password=testpass123" \
  -c cookies.txt

# Use session cookie for authenticated request
curl http://localhost:8080/ -b cookies.txt
```

## Migration from Old System

If you were using the hardcoded `admin/password` account:

1. Run the database migrations
2. Create an admin account via `/register` or SQL
3. All existing user sessions will be invalidated (users need to re-login)
4. Update any scripts or tools that used hardcoded credentials

## Troubleshooting

**"User not found" error:**

- Check that user exists: `SELECT * FROM users WHERE username = 'yourusername';`
- Verify DATABASE_URL is correct

**"Invalid password" error:**

- Password is case-sensitive
- Verify password meets minimum 8 character requirement

**Session not persisting:**

- Check browser accepts cookies
- Verify SECRET_KEY environment variable is set
- Check cookie path is set to `/`

**Registration fails with "username already exists":**

- Username must be unique
- Try a different username or delete the existing user

## Future Enhancements

Potential additions (not currently implemented):

- [ ] Email verification
- [ ] Password reset functionality
- [ ] Account lockout after failed attempts
- [ ] Two-factor authentication (2FA)
- [ ] Role-based access control (RBAC)
- [ ] OAuth integration (Google, GitHub, etc.)
- [ ] Password strength meter
- [ ] Session management (view/revoke active sessions)
- [ ] Account deletion
- [ ] Profile management
