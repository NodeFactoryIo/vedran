package repositories

import (
	"time"

	"github.com/NodeFactoryIo/vedran/internal/models"
	"github.com/asdine/storm/v3"
	"github.com/asdine/storm/v3/q"
)

type RecordRepository interface {
	Save(record *models.Record) error
	// FindSuccessfulRecordsInsideInterval returns all models.Record that happened inside interval
	// defined with arguments from and to
	FindSuccessfulRecordsInsideInterval(nodeID string, from time.Time, to time.Time) ([]models.Record, error)
	CountSuccessfulRequests() (int, error)
	CountFailedRequests() (int, error)
}

type recordRepo struct {
	db *storm.DB
}

func NewRecordRepo(db *storm.DB) RecordRepository {
	return &recordRepo{
		db: db,
	}
}

func (r *recordRepo) Save(record *models.Record) error {
	return r.db.Save(record)
}

func (r *recordRepo) FindSuccessfulRecordsInsideInterval(nodeID string, from time.Time, to time.Time) ([]models.Record, error) {
	var records []models.Record
	err := r.db.Select(q.And(
		q.Eq("NodeId", nodeID),
		q.Gte("Timestamp", from),
		q.Lte("Timestamp", to),
		q.Eq("Status", "successful"),
	)).Find(&records)
	return records, err
}

func (r *recordRepo) CountSuccessfulRequests() (int, error) {
	var records []models.Record
	q := r.db.Select(q.Eq("Status", "successful"))
	err := q.Find(&records)
	if err != nil {
		return 0, err
	}

	return len(records), err
}

func (r *recordRepo) CountFailedRequests() (int, error) {
	var records []models.Record
	q := r.db.Select(q.Eq("Status", "failed"))
	err := q.Find(&records)
	if err != nil {
		return 0, err
	}

	return len(records), err
}
