package repositories

import (
	"github.com/NodeFactoryIo/vedran/internal/models"
	"github.com/asdine/storm/v3"
)

type RecordRepository interface {
	Save(record *models.Record) error
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
