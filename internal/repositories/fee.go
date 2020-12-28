package repositories

import (
	"github.com/NodeFactoryIo/vedran/internal/models"
	"github.com/asdine/storm/v3"
)

type FeeRepository interface {
	RecordNewFee(nodeID string, newFee int64) error
}

type FeeRepo struct {
	db *storm.DB
}

func (f *FeeRepo) RecordNewFee(nodeID string, newFee int64) error {
	feeInDb := &models.Fee{}
	err := f.db.One("NodeId", nodeID, feeInDb)
	if err != nil {
		if err.Error() == "not found" {
			err = f.db.Save(models.Fee{
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
