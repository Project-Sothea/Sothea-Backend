// middleware/tx.go
package middleware

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
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

		serveTransaction(c, tx)
	}
}

// Buffer the response until deferred database constraints have been checked.
type transactionWriter struct {
	gin.ResponseWriter
	header  http.Header
	body    bytes.Buffer
	status  int
	written bool
}

func (w *transactionWriter) Header() http.Header { return w.header }
func (w *transactionWriter) WriteHeader(status int) {
	if !w.written {
		w.status = status
	}
}
func (w *transactionWriter) WriteHeaderNow() { w.written = true }
func (w *transactionWriter) Write(data []byte) (int, error) {
	w.WriteHeaderNow()
	return w.body.Write(data)
}
func (w *transactionWriter) WriteString(data string) (int, error) { return w.Write([]byte(data)) }
func (w *transactionWriter) Status() int                          { return w.status }
func (w *transactionWriter) Size() int {
	if !w.written {
		return -1
	}
	return w.body.Len()
}
func (w *transactionWriter) Written() bool { return w.written }
func (w *transactionWriter) Flush()        { w.WriteHeaderNow() }

func serveTransaction(c *gin.Context, tx pgx.Tx) {
	original := c.Writer
	buffer := &transactionWriter{ResponseWriter: original, header: original.Header().Clone(), status: http.StatusOK}
	c.Writer = buffer
	defer func() { c.Writer = original }()
	c.Next()
	if !c.IsAborted() && buffer.Status() < 400 {
		if err := tx.Commit(c.Request.Context()); err != nil {
			c.Writer = original
			status := http.StatusInternalServerError
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && (pgErr.Code == "23514" || pgErr.Code == "23503" || pgErr.Code == "23505" || pgErr.Code == "P0001") {
				status = http.StatusConflict
			}
			c.AbortWithStatusJSON(status, gin.H{"error": err.Error()})
			return
		}
	}
	for key := range original.Header() {
		original.Header().Del(key)
	}
	for key, values := range buffer.header {
		original.Header()[key] = values
	}
	original.WriteHeader(buffer.status)
	_, _ = original.Write(buffer.body.Bytes())
}

// GetTx extracts the transaction from context (used by repos and tests).
// We store pgx.Tx but keep the return type abstract to avoid import cycles; callers cast to pgx.Tx.
func GetTx(ctx context.Context) (interface{ Rollback(context.Context) error }, bool) {
	tx, ok := ctx.Value(txKey).(interface{ Rollback(context.Context) error })
	return tx, ok
}
