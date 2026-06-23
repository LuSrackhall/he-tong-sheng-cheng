package sqlite

import (
	"asset-leasing-system/internal/domain"
	"time"

	"gorm.io/gorm"
)

func (r *ReceiptRepo) Create(rc *domain.Receipt) error {
	return r.db.Create(rc).Error
}

func (r *ReceiptRepo) GetByID(id uint) (*domain.Receipt, error) {
	var rc domain.Receipt
	err := r.db.First(&rc, id).Error
	if err != nil {
		return nil, err
	}
	return &rc, nil
}

func (r *ReceiptRepo) GetByPaymentID(paymentID uint) (*domain.Receipt, error) {
	var rc domain.Receipt
	err := r.db.Where("payment_id = ?", paymentID).First(&rc).Error
	if err != nil {
		return nil, err
	}
	return &rc, nil
}

func (r *ReceiptRepo) ListByReceiptBookID(bookID uint) ([]domain.Receipt, error) {
	var receipts []domain.Receipt
	err := r.db.Where("receipt_book_id = ?", bookID).Order("sequence_num desc").Find(&receipts).Error
	return receipts, err
}

func (r *ReceiptBookRepo) Create(rb *domain.ReceiptBook) error {
	return r.db.Create(rb).Error
}

func (r *ReceiptBookRepo) GetByID(id uint) (*domain.ReceiptBook, error) {
	var rb domain.ReceiptBook
	err := r.db.First(&rb, id).Error
	if err != nil {
		return nil, err
	}
	return &rb, nil
}

func (r *ReceiptBookRepo) List() ([]domain.ReceiptBook, error) {
	var books []domain.ReceiptBook
	err := r.db.Order("created_at desc").Find(&books).Error
	return books, err
}

func (r *ReceiptBookRepo) Update(rb *domain.ReceiptBook) error {
	return r.db.Save(rb).Error
}

func (r *ReceiptBookRepo) GetActive() (*domain.ReceiptBook, error) {
	var rb domain.ReceiptBook
	err := r.db.Where("status = ?", "active").First(&rb).Error
	if err != nil {
		return nil, err
	}
	return &rb, nil
}

func (r *ReceiptBookRepo) AllocateSequence(bookID uint) (int, error) {
	// 原子递增 CurrentNum，避免并发冲突
	result := r.db.Model(&domain.ReceiptBook{}).
		Where("id = ? AND current_num < start_num + total_pages AND status = ?", bookID, "active").
		Update("current_num", gorm.Expr("current_num + 1"))
	if result.Error != nil {
		return 0, result.Error
	}
	if result.RowsAffected == 0 {
		return 0, nil // 收据本已用完
	}
	var rb domain.ReceiptBook
	if err := r.db.First(&rb, bookID).Error; err != nil {
		return 0, err
	}
	return rb.CurrentNum, nil
}

func (r *TemplateRepo) Create(t *domain.Template) error {
	return r.db.Create(t).Error
}

func (r *TemplateRepo) GetByID(id uint) (*domain.Template, error) {
	var t domain.Template
	err := r.db.First(&t, id).Error
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *TemplateRepo) List() ([]domain.Template, error) {
	var templates []domain.Template
	err := r.db.Order("created_at desc").Find(&templates).Error
	return templates, err
}

func (r *TemplateRepo) Update(t *domain.Template) error {
	return r.db.Save(t).Error
}

func (r *TemplateRepo) Delete(id uint) error {
	return r.db.Delete(&domain.Template{}, id).Error
}

func (r *TemplateRepo) IsUsedByContract(id uint) (bool, error) {
	var count int64
	err := r.db.Model(&domain.Contract{}).Where("template_id = ?", id).Count(&count).Error
	return count > 0, err
}

func (r *UserRepo) Create(u *domain.User) error {
	return r.db.Create(u).Error
}

func (r *UserRepo) GetByUsername(username string) (*domain.User, error) {
	var u domain.User
	err := r.db.Where("username = ?", username).First(&u).Error
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *UserRepo) GetByID(id uint) (*domain.User, error) {
	var u domain.User
	err := r.db.First(&u, id).Error
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *UserRepo) List() ([]domain.User, error) {
	var users []domain.User
	err := r.db.Order("created_at desc").Find(&users).Error
	return users, err
}

func (r *UserRepo) Delete(id uint) error {
	return r.db.Delete(&domain.User{}, id).Error
}

func (r *UserRepo) Count() (int64, error) {
	var count int64
	err := r.db.Model(&domain.User{}).Count(&count).Error
	return count, err
}

func (r *UserRepo) Update(u *domain.User) error {
	return r.db.Save(u).Error
}

func (r *ArrearsRecordRepo) Create(ar *domain.ArrearsRecord) error {
	return r.db.Create(ar).Error
}

func (r *ArrearsRecordRepo) ListByContractID(contractID uint) ([]domain.ArrearsRecord, error) {
	var records []domain.ArrearsRecord
	err := r.db.Where("contract_id = ?", contractID).Order("record_date desc").Find(&records).Error
	return records, err
}

func (r *ArrearsRecordRepo) ListByDateAndLevel(date time.Time, level int) ([]domain.ArrearsRecord, error) {
	var records []domain.ArrearsRecord
	dateStr := date.Format("2006-01-02")
	q := r.db.Where("record_date = ?", dateStr)
	if level > 0 {
		q = q.Where("level = ?", level)
	}
	err := q.Find(&records).Error
	return records, err
}

var _ domain.AssetRepo = (*AssetRepo)(nil)
var _ domain.TenantRepo = (*TenantRepo)(nil)
var _ domain.ContractRepo = (*ContractRepo)(nil)
var _ domain.PaymentRepo = (*PaymentRepo)(nil)
var _ domain.ReceiptRepo = (*ReceiptRepo)(nil)
var _ domain.ReceiptBookRepo = (*ReceiptBookRepo)(nil)
var _ domain.TemplateRepo = (*TemplateRepo)(nil)
var _ domain.UserRepo = (*UserRepo)(nil)
var _ domain.ArrearsRecordRepo = (*ArrearsRecordRepo)(nil)

var _ *gorm.DB = nil
