package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/chonlasit2000/rbac-hexagonal-gorbac/internal/core/domain"
	"github.com/chonlasit2000/rbac-hexagonal-gorbac/internal/core/port"
	"github.com/mikespook/gorbac/v3"
	"github.com/redis/go-redis/v9"
)

type rbacService struct {
	userRepo       port.UserRepository
	roleRepo       port.RoleRepository
	permissionRepo port.PermissionRepository
	rbac           *gorbac.RBAC[string]
	redis          *redis.Client
	mu             sync.RWMutex // ใช้ Lock เวลา Reload Policy
}

func NewRBACService(userRepo port.UserRepository, roleRepo port.RoleRepository, permissionRepo port.PermissionRepository, rdb *redis.Client) port.RBACService {
	return &rbacService{
		userRepo:       userRepo,
		roleRepo:       roleRepo,
		permissionRepo: permissionRepo,
		rbac:           gorbac.New[string](),
		redis:          rdb,
	}
}

func (s *rbacService) LoadPolicy() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.rbac = gorbac.New[string]() // เคลียร์ Policy เก่าออกก่อนโหลดใหม่

	// 1. ดึงข้อมูล Role + Permission จาก Repository
	roles, err := s.roleRepo.GetAll(context.Background())
	if err != nil {
		return err
	}

	// 2. Load เข้า Gorbac
	for _, r := range roles {
		role := gorbac.NewRole(r.Name)
		for _, p := range r.Permissions {
			role.Assign(gorbac.NewPermission(p.Name))
		}
		if err := s.rbac.Add(role); err != nil {
			fmt.Printf("⚠️ Error adding role %s: %v\n", r.Name, err)
		}
	}

	fmt.Printf("✅ RBAC Policy Loaded: %d roles\n", len(roles))
	return nil
}

// CheckAccess แบบมี Redis Cache
func (s *rbacService) CheckAccess(userID string, requiredPerm string) (bool, error) {
	// 1. หาว่า User มี Role อะไรบ้าง (ดึงผ่าน Cache)
	userRoleNames, err := s.getUserRolesWithCache(userID)
	if err != nil {
		return false, err
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	perm := gorbac.NewPermission(requiredPerm)

	// 2. วนลูปเช็คสิทธิ์กับ Gorbac (ใน Memory)
	for _, roleName := range userRoleNames {
		if s.rbac.IsGranted(roleName, perm, nil) {
			return true, nil // ผ่าน
		}
	}

	return false, nil // ไม่ผ่าน
}

// --- Helper: ดึง Role (Redis -> DB fallback) ---
// Best Practice: แยก Logic การดึง Role ออกมาให้ชัดเจน
func (s *rbacService) getUserRolesWithCache(userID string) ([]string, error) {
	ctx := context.Background()
	cacheKey := fmt.Sprintf("rbac:user:%s:roles", userID)

	// A. ลองดึงจาก Redis ก่อน (Fail-safe: ถ้า Redis error ให้ข้ามไป DB เลย)
	val, err := s.redis.Get(ctx, cacheKey).Result()
	if err == nil {
		// Cache HIT!
		var roleNames []string
		if err := json.Unmarshal([]byte(val), &roleNames); err == nil {
			return roleNames, nil
		}
	} else if err != redis.Nil {
		// Redis Error (ไม่ใช่หาไม่เจอ แต่เป็น connection error ฯลฯ)
		log.Printf("⚠️ Redis error: %v (falling back to DB)", err)
	}

	// B. Cache MISS หรือ Redis ล่ม -> ดึงจาก Database
	userRoles, err := s.roleRepo.GetRoleByUserUID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// แปลง Object Role เป็น List of Strings (เพื่อเก็บใน Redis/Gorbac)
	var roleNames []string
	for _, r := range userRoles {
		roleNames = append(roleNames, r.Name)
	}

	// C. บันทึกลง Redis (Background Task)
	// Best Practice: ตั้ง TTL (เช่น 1 ชั่วโมง) เพื่อกันข้อมูลเก่าค้างตลอดกาล
	if len(roleNames) > 0 {
		go func() {
			encoded, _ := json.Marshal(roleNames)
			if err := s.redis.Set(context.Background(), cacheKey, encoded, time.Hour*1).Err(); err != nil {
				log.Printf("⚠️ Failed to set cache: %v", err)
			}
		}()
	}

	return roleNames, nil
}

// 1. สร้าง Role ใหม่
func (s *rbacService) CreateRole(req *port.CreateRoleReq) error {
	role := domain.Role{Name: req.Name}
	if err := s.roleRepo.Create(context.Background(), &role); err != nil {
		return err
	}
	return s.LoadPolicy()
}

// 2. สร้าง Permission ใหม่
func (s *rbacService) CreatePermission(req *port.CreatePermReq) error {
	perm := domain.Permission{Name: req.Name}
	return s.permissionRepo.Create(context.Background(), &perm)
}

// 3. จับคู่ Role <-> Permission
func (s *rbacService) AssignPermissionToRole(req *port.AssignPermReq) error {
	// หา Role และ Permission จาก DB
	role, err := s.roleRepo.GetRoleByName(context.Background(), req.RoleName)
	if err != nil {
		return err
	}

	perm, err := s.permissionRepo.GetPermissionByName(context.Background(), req.PermName)
	if err != nil {
		return err
	}

	// เพิ่มความสัมพันธ์ (GORM Many2Many)
	if err := s.roleRepo.AddAccosiatePermission(context.Background(), role.Uid.String(), perm.Uid.String()); err != nil {
		return err
	}

	// *** สำคัญ: Policy เปลี่ยน ต้องโหลดเข้า Memory ใหม่ ***
	return s.LoadPolicy()
}

// 4. จับคู่ User <-> Role
func (s *rbacService) AssignRoleToUser(req *port.AssignRoleReq) error {
	// หา User และ Role
	user, err := s.userRepo.GetUserByUID(context.Background(), req.UserID)
	if err != nil {
		return err
	}

	role, err := s.roleRepo.GetRoleByName(context.Background(), req.RoleName)
	if err != nil {
		return err
	}

	// เพิ่มความสัมพันธ์
	if err := s.userRepo.AddAccosiateRole(context.Background(), user.Uid.String(), role.Uid.String()); err != nil {
		return err
	}

	cacheKey := fmt.Sprintf("rbac:user:%s:roles", req.UserID)
	s.redis.Del(context.Background(), cacheKey)

	return nil
}

// 1. ยกเลิก Permission ออกจาก Role
func (s *rbacService) RemovePermissionFromRole(req *port.UnassignPermReq) error {
	role, err := s.roleRepo.GetRoleByName(context.Background(), req.RoleName)
	if err != nil {
		return err
	}

	perm, err := s.permissionRepo.GetPermissionByName(context.Background(), req.PermName)
	if err != nil {
		return err
	}

	if err := s.roleRepo.RemoveAssociatePermission(context.Background(), role.Uid.String(), perm.Uid.String()); err != nil {
		return err
	}

	// *** Policy เปลี่ยน ต้อง Reload Gorbac (Memory Cache) ใหม่ ***
	return s.LoadPolicy()
}

// 2. ปลด Role ออกจาก User
func (s *rbacService) RemoveRoleFromUser(req *port.UnassignRoleReq) error {
	user, err := s.userRepo.GetUserByUID(context.Background(), req.UserID)
	if err != nil {
		return err
	}

	role, err := s.roleRepo.GetRoleByName(context.Background(), req.RoleName)
	if err != nil {
		return err
	}

	if err := s.userRepo.RemoveAssociateRole(context.Background(), user.Uid.String(), role.Uid.String()); err != nil {
		return err
	}

	// *** สิทธิ์ของ User คนนี้เปลี่ยน ต้องลบ Cache ทิ้ง (Redis Cache) ***
	cacheKey := fmt.Sprintf("rbac:user:%s:roles", req.UserID)
	s.redis.Del(context.Background(), cacheKey)

	return nil
}

func (s *rbacService) GetAllRoles() ([]domain.Role, error) {
	return s.roleRepo.GetAll(context.Background())
}

func (s *rbacService) GetAllPermissions() ([]domain.Permission, error) {
	return s.permissionRepo.GetAll(context.Background())
}

func (s *rbacService) GetUserRoles(userID string) ([]domain.Role, error) {
	return s.roleRepo.GetRoleByUserUID(context.Background(), userID)
}
