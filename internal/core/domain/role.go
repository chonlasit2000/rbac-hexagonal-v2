package domain

type Role struct {
	Model
	Name string `gorm:"uniqueIndex;not null;size:255" json:"name"`
}
