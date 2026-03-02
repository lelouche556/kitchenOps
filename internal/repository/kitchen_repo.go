package repository

import (
	"errors"
	"sync"

	"gorm.io/gorm"
)

var ErrNotFound = errors.New("not found")

type KitchenRepository struct {
	db *gorm.DB
}

var (
	kitchenRepoOnce sync.Once
	kitchenRepoInst *KitchenRepository
)

func NewKitchenRepository(db *gorm.DB) *KitchenRepository {
	kitchenRepoOnce.Do(func() {
		kitchenRepoInst = &KitchenRepository{db: db}
	})
	return kitchenRepoInst
}

func (r *KitchenRepository) DB() *gorm.DB { return r.db }
