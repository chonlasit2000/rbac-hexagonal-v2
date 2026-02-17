package repository

import (
	"context"

	"github.com/chonlasit2000/rbac-hexagonal-gorbac/internal/core/domain"
	"github.com/chonlasit2000/rbac-hexagonal-gorbac/internal/core/port"
	"gorm.io/gorm"
)

type permissionRepo struct {
	db *gorm.DB
}

func NewPermissionRepository(db *gorm.DB) port.PermissionRepository {
	return &permissionRepo{db: db}
}

func (r *permissionRepo) Create(ctx context.Context, perm *domain.Permission) error {
	return r.db.WithContext(ctx).Create(perm).Error
}

func (r *permissionRepo) GetAll(ctx context.Context) ([]domain.Permission, error) {
	var perms []domain.Permission
	err := r.db.WithContext(ctx).Find(&perms).Error
	if err != nil {
		return nil, err
	}
	return perms, nil
}

func (r *permissionRepo) GetPermissionByName(ctx context.Context, name string) (*domain.Permission, error) {
	var perm domain.Permission
	err := r.db.WithContext(ctx).Where("name = ?", name).First(&perm).Error
	if err != nil {
		return nil, err
	}
	return &perm, nil
}
