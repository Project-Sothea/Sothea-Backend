// middleware/tx.go
package middleware

import (
	"context"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ctxKey struct{}

var txKey ctxKey

func WithTx(db *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		tx, err := db.Begin(c.Request.Context())
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "failed to start tx"})
			return
		}
		defer func() { _ = tx.Rollback(c.Request.Context()) }() // no-op if committed

		// set user for triggers (scoped to THIS tx/session)
		if v, ok := c.Get("userID"); ok {
			if userID, ok := v.(int64); ok {
				if _, err := tx.Exec(
					c.Request.Context(),
					`SELECT set_config('sothea.user_id', $1, true)`,
					strconv.FormatInt(userID, 10),
				); err != nil {
					c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "failed to set user context"})
					return
				}
			}
		}

		// stash tx in ctx
		ctx := context.WithValue(c.Request.Context(), txKey, tx)
		c.Request = c.Request.WithContext(ctx)

		c.Next()

		if c.IsAborted() || c.Writer.Status() >= 400 {
			return // deferred rollback
		}
		_ = tx.Commit(c.Request.Context())
	}
}

// GetTx extracts the transaction from context (used by repos and tests).
// We store pgx.Tx but keep the return type abstract to avoid import cycles; callers cast to pgx.Tx.
func GetTx(ctx context.Context) (interface{ Rollback(context.Context) error }, bool) {
	tx, ok := ctx.Value(txKey).(interface{ Rollback(context.Context) error })
	return tx, ok
}
