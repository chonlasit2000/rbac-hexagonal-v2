package repository

import (
	"context"
	"fmt"

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
	err := r.db.WithContext(ctx).Preload("Permissions").Find(&roles).Error
	if err != nil {
		return nil, err
	}

	return roles, nil
}

func (r *roleRepo) GetRoleByUserUID(ctx context.Context, userUid string) ([]domain.Role, error) {
	var roles []domain.Role
	err := r.db.WithContext(ctx).
		Joins("JOIN user_roles ON user_roles.role_uid = roles.uid"). // ชื่อตารางและคอลัมน์ต้องตรงกับใน DB จริง
		Where("user_roles.user_uid = ?", userUid).
		Find(&roles).Error

	if err != nil {
		return nil, err
	}
	return roles, nil
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

func (r *roleRepo) RemoveAssociatePermission(ctx context.Context, roleID string, permID string) error {
	var role domain.Role
	if err := r.db.WithContext(ctx).Where("uid = ?", roleID).First(&role).Error; err != nil {
		return err
	}
	var perm domain.Permission
	if err := r.db.WithContext(ctx).Where("uid = ?", permID).First(&perm).Error; err != nil {
		return err
	}
	count := r.db.WithContext(ctx).Model(&role).Where("uid = ?", permID).Association("Permissions").Count()
	if count == 0 {
		return fmt.Errorf("role does not have permission: %s", perm.Name)
	}
	// ลบความสัมพันธ์ในตาราง role_permissions
	return r.db.WithContext(ctx).Model(&role).Association("Permissions").Delete(&perm)
}
