package sqlite

import (
	"asset-leasing-system/internal/domain"
)

func (r *PaymentRepo) Create(p *domain.Payment) error {
	return r.db.Create(p).Error
}

func (r *PaymentRepo) ListByContractID(contractID uint) ([]domain.Payment, error) {
	var payments []domain.Payment
	err := r.db.Where("contract_id = ?", contractID).Order("paid_at desc").Find(&payments).Error
	return payments, err
}
