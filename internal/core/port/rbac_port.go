package port

type RBACService interface {
	LoadPolicy() error
	CheckAccess(userID string, requiredPerm string) (bool, error)

	// --- CRUD Methods ---
	CreateRole(req *CreateRoleReq) error
	CreatePermission(req *CreatePermReq) error
	AssignPermissionToRole(req *AssignPermReq) error
	AssignRoleToUser(req *AssignRoleReq) error
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
