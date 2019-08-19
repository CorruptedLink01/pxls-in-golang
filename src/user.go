package main

// UserRole is the role of an user
type UserRole string

const (
	// DefaultUserRole is the role an user which can only
	// place pixels, lookup pixels, report other users, and chat
	DefaultUserRole = "USER"
)

// AuthData represents the authentication data of an user
type AuthData struct {
	Method string
	IP     string
	Token  string
}

// User represents an user placing on the canvas
type User struct {
	ID           uint
	Name         string
	Role         UserRole
	Auth         AuthData
	PixelStacker *PixelStacker
}

// MakeUser creates an User with the defined id, name and role
func MakeUser(id uint, name string, role UserRole) *User {
	u := &User{
		ID:   id,
		Name: name,
		Role: role,
	}
	u.PixelStacker = MakePixelStacker()
	return u
}

// MakeUserFromIP creates an User with the authentication data set to IP
func MakeUserFromIP(id uint, ip string) *User {
	u := MakeUser(id, "-snip-", DefaultUserRole)
	u.Auth = AuthData{
		Method: "IP",
		IP:     ip,
		Token:  "ip:" + ip,
	}
	return u
}
