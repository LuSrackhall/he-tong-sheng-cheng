package sqlite

import (
	"asset-leasing-system/internal/domain"
)

func (r *PaymentRepo) Create(p *domain.Payment) error {
	return r.db.Create(p).Error
}

func (r *PaymentRepo) GetByID(id uint) (*domain.Payment, error) {
	var p domain.Payment
	err := r.db.First(&p, id).Error
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *PaymentRepo) ListByContractID(contractID uint) ([]domain.Payment, error) {
	var payments []domain.Payment
	err := r.db.Where("contract_id = ?", contractID).Order("paid_at desc").Find(&payments).Error
	return payments, err
}
