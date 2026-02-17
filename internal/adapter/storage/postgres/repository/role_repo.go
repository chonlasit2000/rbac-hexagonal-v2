package repository

import (
	"context"

	"github.com/chonlasit2000/rbac-hexagonal-gorbac/internal/core/domain"
	"github.com/chonlasit2000/rbac-hexagonal-gorbac/internal/core/port"
	"gorm.io/gorm"
)

type roleRepo struct {
	db *gorm.DB
}

func NewRoleRepository(db *gorm.DB) port.RoleRepository {
	return &roleRepo{db: db}
}

func (r *roleRepo) Create(ctx context.Context, role *domain.Role) error {
	return r.db.WithContext(ctx).Create(role).Error
}

func (r *roleRepo) GetAll(ctx context.Context) ([]domain.Role, error) {
	var roles []domain.Role
	err := r.db.WithContext(ctx).Preload("Permission").Find(&roles).Error
	if err != nil {
		return nil, err
	}

	return roles, nil
}

func (r *roleRepo) GetRoleByUserUID(ctx context.Context, userUid string) ([]domain.Role, error) {
	var role []domain.Role
	err := r.db.WithContext(ctx).Where("user_uid = ?", userUid).First(&role).Error
	if err != nil {
		return nil, err
	}
	return role, nil
}

func (r *roleRepo) GetRoleByName(ctx context.Context, name string) (*domain.Role, error) {
	var role domain.Role
	err := r.db.WithContext(ctx).Where("name = ?", name).First(&role).Error
	if err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *roleRepo) AddAccosiatePermission(ctx context.Context, roleID string, permID string) error {
	// หา Role และ Permission
	var role domain.Role
	if err := r.db.WithContext(ctx).Where("uid = ?", roleID).First(&role).Error; err != nil {
		return err
	}
	var perm domain.Permission
	if err := r.db.WithContext(ctx).Where("uid = ?", permID).First(&perm).Error; err != nil {
		return err
	}
	// จับคู่ Role <-> Permission
	return r.db.WithContext(ctx).Model(&role).Association("Permissions").Append(&perm)
}
