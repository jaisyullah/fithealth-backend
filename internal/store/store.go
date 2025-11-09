package store

import (
	"context"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type DB struct {
	*gorm.DB
}

type ObservationRaw struct {
	ID          int64     `gorm:"primaryKey;autoIncrement"`
	DeviceID    string    `gorm:"type:text"`
	PatientID   string    `gorm:"type:text"`
	ObsType     string    `gorm:"type:text"`
	Value       float64
	Unit        string    `gorm:"type:text"`
	ObservedAt  time.Time `gorm:"type:timestamptz"`
	ReceivedAt  time.Time `gorm:"autoCreateTime"`
	Status      string    `gorm:"type:text"` // pending, queued, sending, sent, failed
	RetryCount  int
	LastError   string    `gorm:"type:text"`
}

type FHIRTransaction struct {
	ID               int64     `gorm:"primaryKey;autoIncrement"`
	ObservationRawID int64
	FHIRPayload      string    `gorm:"type:jsonb"`
	ResponseCode     int
	ResponseBody     string    `gorm:"type:text"`
	SentAt           time.Time `gorm:"autoCreateTime"`
	Status           string    `gorm:"type:text"` // success, failed
}

func NewGorm(dsn string) (*DB, error) {
	gdb, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	db := &DB{DB: gdb}
	if err := db.AutoMigrate(&ObservationRaw{}, &FHIRTransaction{}); err != nil {
		return nil, err
	}
	return db, nil
}

// Create incoming raw observation
func (db *DB) CreateObservation(ctx context.Context, o *ObservationRaw) error {
	return db.WithContext(ctx).Create(o).Error
}

func (db *DB) GetObservationByID(ctx context.Context, id int64) (*ObservationRaw, error) {
	var o ObservationRaw
	if err := db.WithContext(ctx).First(&o, id).Error; err != nil {
		return nil, err
	}
	return &o, nil
}

func (db *DB) GetPendingObservations(ctx context.Context, limit int) ([]ObservationRaw, error) {
	var out []ObservationRaw
	if err := db.WithContext(ctx).Where("status IN ?", []string{"pending","queued","failed"}).Limit(limit).Find(&out).Error; err != nil {
		return nil, err
	}
	return out, nil
}

func (db *DB) UpdateStatus(ctx context.Context, id int64, status string, lastErr string) error {
	return db.WithContext(ctx).Model(&ObservationRaw{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":     status,
		"last_error": lastErr,
	}).Error
}

func (db *DB) IncrementRetry(ctx context.Context, id int64, lastErr string) error {
	var o ObservationRaw
	if err := db.WithContext(ctx).First(&o, id).Error; err != nil {
		return err
	}
	o.RetryCount += 1
	o.LastError = lastErr
	if o.RetryCount > 5 {
		o.Status = "failed"
	} else {
		o.Status = "queued"
	}
	return db.WithContext(ctx).Save(&o).Error
}

func (db *DB) MarkSent(ctx context.Context, id int64, respCode int) error {
	return db.WithContext(ctx).Model(&ObservationRaw{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status": "sent",
	}).Error
}

func (db *DB) SaveFHIRTransaction(ctx context.Context, t *FHIRTransaction) error {
	return db.WithContext(ctx).Create(t).Error
}

// helper to parse and validate
func (db *DB) VerifyConnection() error {
	sqlDB, err := db.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Ping()
}
