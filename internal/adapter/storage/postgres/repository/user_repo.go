package repository

import (
	"context"
	"fmt"

	"github.com/chonlasit2000/rbac-hexagonal-gorbac/internal/core/domain"
	"github.com/chonlasit2000/rbac-hexagonal-gorbac/internal/core/port"
	"gorm.io/gorm"
)

type userRepo struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) port.UserRepository {
	return &userRepo{db: db}
}

func (r *userRepo) Create(ctx context.Context, user *domain.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

func (r *userRepo) GetUserByUID(ctx context.Context, uid string) (*domain.User, error) {
	var user domain.User
	err := r.db.WithContext(ctx).Where("uid = ?", uid).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepo) GetUserByUsername(ctx context.Context, username string) (*domain.User, error) {
	var user domain.User
	err := r.db.WithContext(ctx).Where("username = ?", username).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepo) AddAccosiateRole(ctx context.Context, userID string, roleID string) error {
	// หา User และ Role
	var user domain.User
	if err := r.db.WithContext(ctx).Where("uid = ?", userID).First(&user).Error; err != nil {
		return err
	}
	var role domain.Role
	if err := r.db.WithContext(ctx).Where("uid = ?", roleID).First(&role).Error; err != nil {
		return err
	}
	// จับคู่ User <-> Role
	return r.db.WithContext(ctx).Model(&user).Association("Roles").Append(&role)
}

func (r *userRepo) RemoveAssociateRole(ctx context.Context, userID string, roleID string) error {
	var user domain.User
	if err := r.db.WithContext(ctx).Where("uid = ?", userID).First(&user).Error; err != nil {
		return err
	}
	var role domain.Role
	if err := r.db.WithContext(ctx).Where("uid = ?", roleID).First(&role).Error; err != nil {
		return err
	}
	count := r.db.WithContext(ctx).Model(&user).Where("uid = ?", roleID).Association("Roles").Count()
	if count == 0 {
		return fmt.Errorf("user does not have role: %s", role.Name)
	}
	// ลบความสัมพันธ์ในตาราง user_roles
	return r.db.WithContext(ctx).Model(&user).Association("Roles").Delete(&role)
}
