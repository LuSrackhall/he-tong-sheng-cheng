package sqlite

import (
	"asset-leasing-system/internal/domain"
)

func (r *ContractRepo) Create(c *domain.Contract) error {
	return r.db.Create(c).Error
}

func (r *ContractRepo) GetByID(id uint) (*domain.Contract, error) {
	var c domain.Contract
	err := r.db.Preload("Asset").Preload("Tenant").First(&c, id).Error
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *ContractRepo) List(search string, status string, offset, limit int) ([]domain.Contract, int64, error) {
	var contracts []domain.Contract
	var total int64
	q := r.db.Model(&domain.Contract{}).Preload("Asset").Preload("Tenant")
	if search != "" {
		q = q.Joins("JOIN tenants ON tenants.id = contracts.tenant_id").
			Joins("JOIN assets ON assets.id = contracts.asset_id").
			Where("tenants.name LIKE ? OR assets.name LIKE ?", "%"+search+"%", "%"+search+"%")
	}
	if status != "" {
		q = q.Where("contracts.status = ?", status)
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := q.Offset(offset).Limit(limit).Order("contracts.created_at desc").Find(&contracts).Error; err != nil {
		return nil, 0, err
	}
	return contracts, total, nil
}

func (r *ContractRepo) Update(c *domain.Contract) error {
	return r.db.Save(c).Error
}

func (r *ContractRepo) ListByAssetID(assetID uint) ([]domain.Contract, error) {
	var contracts []domain.Contract
	err := r.db.Where("asset_id = ?", assetID).Order("created_at desc").Find(&contracts).Error
	return contracts, err
}

func (r *ContractRepo) ListByTenantID(tenantID uint) ([]domain.Contract, error) {
	var contracts []domain.Contract
	err := r.db.Where("tenant_id = ?", tenantID).Order("created_at desc").Find(&contracts).Error
	return contracts, err
}

func (r *ContractRepo) ListActive() ([]domain.Contract, error) {
	var contracts []domain.Contract
	err := r.db.Where("status IN ?", []string{"active", "arrears"}).Find(&contracts).Error
	return contracts, err
}

func (r *ContractRepo) ListUnpaid() ([]domain.Contract, error) {
	var contracts []domain.Contract
	err := r.db.Where("status != ?", "paidup").Preload("Asset").Preload("Tenant").Find(&contracts).Error
	return contracts, err
}
