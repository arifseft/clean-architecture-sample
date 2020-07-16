package user

import "time"

type User struct {
	ID          uint64     `gorm:"primary_key" json:"id"`
	Email       string     `json:"email,omitempty"`
	FirstName   string     `json:"first_name,omitempty"`
	LastName    string     `json:"last_name,omitempty"`
	Password    string     `json:"password,omitempty"`
	PhoneNumber *string    `json:"phone_number,omitempty"`
	Address     *string    `json:"address,omitempty"`
	DisplayPic  *string    `json:"display_pic,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	DeletedAt   *time.Time `sql:"index" json:"deleted_at"`
}
