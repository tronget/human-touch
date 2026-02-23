package user

type User struct {
	ID       int64  `db:"id" json:"id"`
	Email    string `db:"email" json:"email"`
	Password string `db:"password" json:"-"`
	Name     string `db:"name" json:"name"`
}

func (u *User) GetName() string {
	return u.Name
}

func (u *User) GetID() int64 {
	return u.ID
}

func NewUser(email, password, name string) *User {
	return &User{
		Email:    email,
		Password: password,
		Name:     name,
	}
}
