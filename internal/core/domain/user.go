package domain

type User struct {
	Model
	Username string `gorm:"uniqueIndex;not null;size:255" json:"username"`
	Email    string `gorm:"uniqueIndex;not null;size:255" json:"email"`
	Password string `gorm:"size:255" json:"password"`
}
