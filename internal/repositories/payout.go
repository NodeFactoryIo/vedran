package repositories

import (
	"github.com/NodeFactoryIo/vedran/internal/models"
	"github.com/asdine/storm/v3"
)

type PayoutRepository interface {
	Save(payment *models.Payout) error
	GetAll() (*[]models.Payout, error)
	FindLatestPayout() (*models.Payout, error)
}

type payoutRepo struct {
	db *storm.DB
}

func NewPayoutRepo(db *storm.DB) PayoutRepository {
	return &payoutRepo{
		db: db,
	}
}

func (p *payoutRepo) Save(payment *models.Payout) error {
	return p.db.Save(payment)
}

func (p *payoutRepo) GetAll() (*[]models.Payout, error) {
	var payouts []models.Payout
	err := p.db.All(&payouts)
	return &payouts, err
}

func (p *payoutRepo) FindLatestPayout() (*models.Payout, error) {
	var payout models.Payout
	err := p.db.Select().OrderBy("Timestamp").Reverse().First(&payout)
	return &payout, err
}
