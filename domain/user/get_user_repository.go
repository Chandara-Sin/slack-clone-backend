package user

import "gorm.io/gorm"

func GetUser(db *gorm.DB) func(string) (User, error) {
	return func(ID string) (User, error) {
		usr := User{}
		r := db.First(&usr, ID)
		return usr, r.Error
	}
}

func GetUserByEmail(db *gorm.DB) func(string) (User, error) {
	return func(eml string) (User, error) {
		usr := User{}
		err := db.Where("email = ?", eml).First(&usr).Error
		return usr, err
	}
}
