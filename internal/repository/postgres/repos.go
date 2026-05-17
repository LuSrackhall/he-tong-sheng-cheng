package postgres

import (
	"asset-leasing-system/internal/domain"
)

func (r *AssetRepo) Create(a *domain.Asset) error {
	return r.db.Create(a).Error
}

func (r *AssetRepo) GetByID(id uint) (*domain.Asset, error) {
	var a domain.Asset
	err := r.db.First(&a, id).Error
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (r *AssetRepo) List(search string, assetType string, offset, limit int) ([]domain.Asset, int64, error) {
	var assets []domain.Asset
	var total int64
	q := r.db.Model(&domain.Asset{})
	if search != "" {
		q = q.Where("name LIKE ? OR description LIKE ?", "%"+search+"%", "%"+search+"%")
	}
	if assetType != "" {
		q = q.Where("asset_type = ?", assetType)
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := q.Offset(offset).Limit(limit).Order("created_at desc").Find(&assets).Error; err != nil {
		return nil, 0, err
	}
	return assets, total, nil
}

func (r *AssetRepo) Update(a *domain.Asset) error {
	return r.db.Save(a).Error
}

func (r *TenantRepo) Create(t *domain.Tenant) error {
	return r.db.Create(t).Error
}

func (r *TenantRepo) GetByID(id uint) (*domain.Tenant, error) {
	var t domain.Tenant
	err := r.db.First(&t, id).Error
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *TenantRepo) List(search string, offset, limit int) ([]domain.Tenant, int64, error) {
	var tenants []domain.Tenant
	var total int64
	q := r.db.Model(&domain.Tenant{})
	if search != "" {
		q = q.Where("name LIKE ? OR phone LIKE ? OR id_card LIKE ?", "%"+search+"%", "%"+search+"%", "%"+search+"%")
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := q.Offset(offset).Limit(limit).Order("created_at desc").Find(&tenants).Error; err != nil {
		return nil, 0, err
	}
	return tenants, total, nil
}

func (r *TenantRepo) Update(t *domain.Tenant) error {
	return r.db.Save(t).Error
}
