package model

// User gophkeeper user
type User struct {
	// ID user identifier
	ID int64
	// Login user login
	Login string
	// HashedPassword user hashed password
	HashedPassword string
	// Timestamp of last modification of user
	Timestamp int64
}

// EqualTo returns users equality
func (u *User) EqualTo(a User) bool {
	return u.ID == a.ID && u.Login == a.Login && u.HashedPassword == a.HashedPassword && u.Timestamp == a.Timestamp
}
