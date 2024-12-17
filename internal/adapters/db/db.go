package db

import (
	"context"
	"fmt"

	"github.com/chyiyaqing/gmicro-payment/internal/application/core/domain"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Adapter struct {
	db *gorm.DB
}

func NewAdapter(sqliteDB string) (*Adapter, error) {
	db, openErr := gorm.Open(sqlite.Open(sqliteDB), &gorm.Config{})
	if openErr != nil {
		return nil, fmt.Errorf("db open %v error: %v", sqliteDB, openErr)
	}
	err := db.AutoMigrate(&Payment{})
	if err != nil {
		return nil, fmt.Errorf("db migrate error: %v", err)
	}
	return &Adapter{db: db}, nil
}

type Payment struct {
	gorm.Model
	CustomerID int64
	Status     string
	OrderID    int64
	TotalPrice float32
}

func (a Adapter) Get(ctx context.Context, id string) (domain.Payment, error) {
	var paymentEntity Payment
	res := a.db.WithContext(ctx).First(&paymentEntity, id)
	payment := domain.Payment{
		ID:         int64(paymentEntity.ID),
		CustomerID: paymentEntity.CustomerID,
		Status:     paymentEntity.Status,
		OrderId:    paymentEntity.OrderID,
		TotalPrice: paymentEntity.TotalPrice,
		CreatedAt:  paymentEntity.CreatedAt.UnixNano(),
	}
	return payment, res.Error
}

func (a Adapter) Save(ctx context.Context, payment *domain.Payment) error {
	orderModel := Payment{
		CustomerID: payment.CustomerID,
		Status:     payment.Status,
		OrderID:    payment.OrderId,
		TotalPrice: payment.TotalPrice,
	}
	res := a.db.WithContext(ctx).Create(&orderModel)
	if res.Error == nil {
		payment.ID = int64(orderModel.ID)
	}
	return res.Error
}
