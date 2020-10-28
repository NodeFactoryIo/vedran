package repositories

import (
	"github.com/NodeFactoryIo/vedran/internal/models"
	"github.com/asdine/storm/v3"
	"github.com/asdine/storm/v3/q"
	"time"
)

type RecordRepository interface {
	Save(record *models.Record) error
	// FindSuccessfulRecordsInsideInterval returns all models.Record that happened inside interval
	// defined with arguments from and to
	FindSuccessfulRecordsInsideInterval(nodeID string, from time.Time, to time.Time) ([]models.Record, error)
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
		q.Eq("id", nodeID),
		q.Gte("timestamp", from),
		q.Lte("timestamp", to),
		q.Eq("status", "successful"),
	)).Find(&records)
	return records, err
}
