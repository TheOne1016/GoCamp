package repository

import (
	"context"
	"GoCamp/webook/internal/repository/cache"
)

type CodeRepository struct {
	cache *cache.CodeCache
}

func NewCodeRepository() *CodeRepository {
	return &CodeRepository{
		cache:
	}
}

func (repo *CodeRepository) Store(ctx context.Context, biz, phone, code string) error {
	return repo.cache.Set(ctx, biz, phone, code)
}