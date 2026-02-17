package domain

type Role struct {
	Model
	Name string `gorm:"uniqueIndex;not null;size:255" json:"name"`

	// เพิ่มความสัมพันธ์
	Permissions []*Permission `gorm:"many2many:role_permissions;" json:"permissions"`
	Users       []*User       `gorm:"many2many:user_roles;" json:"-"`
}
