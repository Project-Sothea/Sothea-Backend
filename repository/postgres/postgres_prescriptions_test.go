package postgres

import (
	"context"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"sothea-backend/entities"
	db "sothea-backend/repository/sqlc"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Set TEST_DATABASE_URL to a disposable database initialized with sql/1..4.
func TestConcurrentPrescriptionCreation(t *testing.T) {
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		t.Skip("TEST_DATABASE_URL is not set")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	pool, err := pgxpool.New(ctx, url)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	var patientID int64
	// Seed IDs may exceed the sequence, so choose a free ID explicitly.
	err = pool.QueryRow(ctx, `INSERT INTO patient_details(id, name, family_group, khmer_name, dob, gender, village)
		SELECT COALESCE(MAX(id),0)+1, 'concurrent rx test', 'test', 'test', '2000-01-01', 'M', 'test'
		FROM patient_details RETURNING id`).Scan(&patientID)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _, _ = pool.Exec(context.Background(), "DELETE FROM patient_details WHERE id=$1", patientID) }()
	_, err = pool.Exec(ctx, `INSERT INTO admin(id, vid, reg_date, queue_no, pregnant, sent_to_id) VALUES($1,1,CURRENT_DATE,'test',FALSE,FALSE)`, patientID)
	if err != nil {
		t.Fatal(err)
	}
	repo := NewPostgresPrescriptionRepository(pool)
	const clients = 8
	results := make(chan *entities.Prescription, clients)
	errors := make(chan error, clients)
	start := make(chan struct{})
	var group sync.WaitGroup
	for i := 0; i < clients; i++ {
		group.Add(1)
		go func(i int) {
			defer group.Done()
			<-start
			notes := fmt.Sprintf("client %d", i)
			p, err := repo.CreatePrescription(ctx, &entities.Prescription{Prescription: db.Prescription{PatientID: int32(patientID), Vid: 1, Notes: &notes}})
			if err != nil {
				errors <- err
				return
			}
			results <- p
		}(i)
	}
	close(start)
	group.Wait()
	close(results)
	close(errors)
	for err := range errors {
		t.Errorf("create failed: %v", err)
	}
	var first *entities.Prescription
	count := 0
	for result := range results {
		count++
		if first == nil {
			first = result
		}
		if result.ID != first.ID || *result.Notes != *first.Notes {
			t.Error("concurrent clients received different headers")
		}
	}
	if count != clients {
		t.Fatalf("received %d headers, want %d", count, clients)
	}
}
