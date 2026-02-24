package port

import "github.com/chonlasit2000/rbac-hexagonal-gorbac/internal/core/domain"

type RBACService interface {
	LoadPolicy() error
	CheckAccess(userID string, requiredPerm string) (bool, error)

	// --- CRUD Methods ---
	CreateRole(req *CreateRoleReq) error
	CreatePermission(req *CreatePermReq) error
	AssignPermissionToRole(req *AssignPermReq) error
	AssignRoleToUser(req *AssignRoleReq) error
	RemovePermissionFromRole(req *UnassignPermReq) error
	RemoveRoleFromUser(req *UnassignRoleReq) error

	GetAllRoles() ([]domain.Role, error)
	GetAllPermissions() ([]domain.Permission, error)
	GetUserRoles(userID string) ([]domain.Role, error)
}

type CreateRoleReq struct {
	Name string `json:"name"`
}

type CreatePermReq struct {
	Name string `json:"name"`
}

type AssignPermReq struct {
	RoleName string `json:"role_name"`
	PermName string `json:"perm_name"`
}

type AssignRoleReq struct {
	UserID   string `json:"user_id"`
	RoleName string `json:"role_name"`
}

type UnassignPermReq struct {
	RoleName string `json:"role_name"`
	PermName string `json:"perm_name"`
}

type UnassignRoleReq struct {
	UserID   string `json:"user_id"`
	RoleName string `json:"role_name"`
}
