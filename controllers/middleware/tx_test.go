package middleware

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type commitTx struct {
	pgx.Tx
	commit func() error
}

func TestDeferredAllocationFailure(t *testing.T) {
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		t.Skip("TEST_DATABASE_URL is not set")
	}
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, url)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	tx, err := pool.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	// The batch is deliberately for a different drug. The FK is valid; the
	// drug-match constraint fires only at commit, after the handler writes JSON.
	_, err = tx.Exec(ctx, `
		INSERT INTO patient_details(id, name, family_group, khmer_name, dob, gender, village)
		VALUES(900002,'deferred test','test','test','2000-01-01','M','test');
		INSERT INTO admin(id,vid,reg_date,queue_no,pregnant,sent_to_id)
		VALUES(900002,1,CURRENT_DATE,'test',FALSE,FALSE);
		DO $$ DECLARE rx BIGINT; line BIGINT; drug BIGINT; other_drug BIGINT; batch BIGINT; loc BIGINT;
		BEGIN
		  INSERT INTO drugs(generic_name,dosage_form_code,route_code,dispense_unit)
		  VALUES('deferred test','TAB','PO','tab') RETURNING id INTO drug;
		  INSERT INTO drugs(generic_name,dosage_form_code,route_code,dispense_unit)
		  VALUES('deferred wrong drug','TAB','PO','tab') RETURNING id INTO other_drug;
		  INSERT INTO drug_batches(drug_id,batch_number) VALUES(drug,'test') RETURNING id INTO batch;
		  INSERT INTO batch_locations(batch_id,location,quantity) VALUES(batch,'test',10);
		  INSERT INTO drug_batches(drug_id,batch_number) VALUES(other_drug,'test') RETURNING id INTO batch;
		  INSERT INTO batch_locations(batch_id,location,quantity) VALUES(batch,'test',10) RETURNING id INTO loc;
		  INSERT INTO prescriptions(patient_id,vid) VALUES(900002,1) RETURNING id INTO rx;
		  INSERT INTO prescription_lines(prescription_id,drug_id,dose_amount,dose_unit,frequency_code,duration,duration_unit)
		  VALUES(rx,drug,1,'tab','OM',1,'day') RETURNING id INTO line;
		  INSERT INTO prescription_batch_items(line_id,batch_location_id,quantity) VALUES(line,loc,1);
		END $$;
	`)
	if err != nil {
		t.Fatal(err)
	}
	recorder := httptest.NewRecorder()
	router := gin.New()
	router.Use(func(c *gin.Context) { serveTransaction(c, tx) })
	router.GET("/", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"saved": true}) })
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/", nil))
	if recorder.Code != http.StatusConflict || strings.Contains(recorder.Body.String(), "saved") {
		t.Fatalf("unexpected response: %d %s", recorder.Code, recorder.Body.String())
	}
	var exists bool
	if err := pool.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM patient_details WHERE id=900002)").Scan(&exists); err != nil {
		t.Fatal(err)
	}
	if exists {
		t.Fatal("rejected transaction committed fixture data")
	}
}

func (tx commitTx) Commit(context.Context) error { return tx.commit() }

func TestTransactionResponse(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, test := range []struct {
		name          string
		commitErr     error
		handlerStatus int
		wantStatus    int
		wantCommit    bool
	}{
		{"committed", nil, 201, 201, true},
		{"deferred constraint", &pgconn.PgError{Code: "23514", Message: "allocation rejected"}, 201, 409, true},
		{"commit failure", errors.New("connection lost"), 200, 500, true},
		{"handler failure", nil, 400, 400, false},
	} {
		t.Run(test.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			committed := false
			tx := commitTx{commit: func() error {
				if recorder.Body.Len() != 0 || recorder.Header().Get("X-Success") != "" {
					t.Fatal("response reached client before commit")
				}
				committed = true
				return test.commitErr
			}}
			router := gin.New()
			router.Use(func(c *gin.Context) { serveTransaction(c, tx) })
			router.GET("/", func(c *gin.Context) {
				c.Header("X-Success", "yes")
				c.JSON(test.handlerStatus, gin.H{"saved": true})
			})
			router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/", nil))
			if recorder.Code != test.wantStatus || committed != test.wantCommit {
				t.Fatalf("status=%d committed=%v", recorder.Code, committed)
			}
			if test.commitErr != nil {
				if strings.Contains(recorder.Body.String(), "saved") || recorder.Header().Get("X-Success") != "" {
					t.Fatal("commit failure leaked success response")
				}
			} else if !strings.Contains(recorder.Body.String(), "saved") {
				t.Fatal("response lost")
			}
		})
	}
}
