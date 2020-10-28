package repositories

import (
	"github.com/NodeFactoryIo/vedran/internal/models"
	"github.com/asdine/storm/v3"
)

type PaymentRepository interface {
	Save(payment *models.Payment) error
}

type paymentRepo struct {
	db *storm.DB
}

func NewPaymentRepo(db *storm.DB) PaymentRepository {
	return &paymentRepo{
		db: db,
	}
}

func (p paymentRepo) Save(payment *models.Payment) error {
	return p.db.Save(payment)
}
