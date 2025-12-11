#!/usr/bin/env bash
# Create an initial admin user for vidserve5
# Usage: ./create_admin.sh [username] [password]

set -e

USERNAME="${1:-admin}"
PASSWORD="${2:-changeme123}"

echo "Creating admin user: $USERNAME"

# Generate password hash using Go
HASH=$(cat <<EOF | go run -
package main
import (
    "fmt"
    "golang.org/x/crypto/bcrypt"
    "os"
)
func main() {
    hash, err := bcrypt.GenerateFromPassword([]byte(os.Args[1]), bcrypt.DefaultCost)
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        os.Exit(1)
    }
    fmt.Print(string(hash))
}
EOF
"$PASSWORD"
)

if [ $? -ne 0 ]; then
    echo "Failed to generate password hash"
    exit 1
fi

# Insert user into database
psql whutbot <<SQL
INSERT INTO users (id, username, password_hash)
VALUES (gen_random_uuid(), '$USERNAME', '$HASH')
ON CONFLICT (username) DO UPDATE 
SET password_hash = EXCLUDED.password_hash;

SELECT id, username, created_at FROM users WHERE username = '$USERNAME';
SQL

echo ""
echo "✓ User created successfully!"
echo "  Username: $USERNAME"
echo "  Password: $PASSWORD"
echo ""
echo "IMPORTANT: Change the default password after first login!"
