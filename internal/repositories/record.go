package repositories

import (
	"github.com/NodeFactoryIo/vedran/internal/models"
	"github.com/asdine/storm/v3"
)

type RecordRepo struct {
	db *storm.DB
}

func NewRecordRepo(db *storm.DB) *RecordRepo {
	return &RecordRepo{
		db: db,
	}
}

func (r *RecordRepo) Save(record *models.Record) error {
	return r.db.Save(record)
}
