package creators

import "github.com/jackc/pgx/v5/pgxpool"

var dbPool *pgxpool.Pool

func SetDB(pool *pgxpool.Pool) {
	dbPool = pool
}
