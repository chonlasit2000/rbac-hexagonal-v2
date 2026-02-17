package domain

type User struct {
	Model
	Username string `gorm:"uniqueIndex;not null;size:255" json:"username"`
	Email    string `gorm:"uniqueIndex;not null;size:255" json:"email"`
	Password string `gorm:"size:255" json:"password"`

	// เพิ่มบรรทัดนี้: User มีได้หลาย Role
	Roles []*Role `gorm:"many2many:user_roles;" json:"roles"`
}
