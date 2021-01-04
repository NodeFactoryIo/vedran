package repositories

import (
	"github.com/NodeFactoryIo/vedran/internal/models"
	"github.com/asdine/storm/v3"
)

type FeeRepository interface {
	RecordNewFee(nodeID string, newFee int64) error
	GetAllFees() (*[]models.Fee, error)
}

type feeRepo struct {
	db *storm.DB
}

func NewFeeRepo(db *storm.DB) FeeRepository {
	return &feeRepo{
		db: db,
	}
}

func (f *feeRepo) RecordNewFee(nodeID string, newFee int64) error {
	feeInDb := &models.Fee{}
	err := f.db.One("NodeId", nodeID, feeInDb)
	if err != nil {
		if err.Error() == "not found" {
			err = f.db.Save(&models.Fee{
				NodeId:   nodeID,
				TotalFee: newFee,
			})
		}
		return err
	}

	feeInDb.TotalFee += newFee
	err = f.db.Update(feeInDb)
	return err
}

func (f *feeRepo) GetAllFees() (*[]models.Fee, error) {
	var fees []models.Fee
	err := f.db.All(&fees)
	return &fees, err
}
