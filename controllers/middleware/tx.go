// middleware/tx.go
package middleware

import (
	"context"
	"database/sql"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type ctxKey struct{}

var txKey ctxKey

func WithTx(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		tx, err := db.BeginTx(c.Request.Context(), nil)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "failed to start tx"})
			return
		}
		defer tx.Rollback() // no-op if committed

		// set user for triggers (scoped to THIS tx/session)
		if v, ok := c.Get("userID"); ok {
			if userID, ok := v.(int64); ok {
				if _, err := tx.ExecContext(
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
		_ = tx.Commit()
	}
}

// Getter (used by repos)
func GetTx(ctx context.Context) (*sql.Tx, bool) {
	tx, ok := ctx.Value(txKey).(*sql.Tx)
	return tx, ok
}

// used for tests
func CtxWithTx(ctx context.Context, tx *sql.Tx) context.Context {
	return context.WithValue(ctx, txKey, tx)
}
