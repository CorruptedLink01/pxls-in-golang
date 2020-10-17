package main

import (
	"fmt"
	"strings"
)

// UserRole is the role of an user
type UserRole string

const (
	// DefaultUserRole is the role an user which can only
	// place pixels, lookup pixels, report other users, and chat
	DefaultUserRole = "USER"
)

// UserLogin holds user login information.
type UserLogin struct {
	Method string
	// TODO(netux): figure out a better name than "Foo" for Service ID or IP
	Foo string
}

func (login *UserLogin) String() string {
	return login.Method + ":" + login.Foo
}

// ParseUserLogin parses a raw login string in the form "{Method}:{Foo}"
// into an UserLogin.
func ParseUserLogin(login string) UserLogin {
	s := strings.SplitN(login, ":", 2)
	return UserLogin{
		Method: s[0],
		Foo:    s[1],
	}
}

// User represents an user placing on the canvas
type User struct {
	*DBUser
	PixelStacker *PixelStacker
}

// MakeUser creates an User with the given DBUser
func MakeUser(dbUser *DBUser) *User {
	u := &User{
		dbUser,
		MakePixelStacker(),
	}
	return u
}

// UserList contains cached users stored by different criteria.
type UserList struct {
	byID        map[uint]*User
	byTokenOrIP map[string]*User
}

// GetByID returns a cached user searched by it's ID.
func (l *UserList) GetByID(id uint) (u *User, ok bool) {
	u, ok = l.byID[id]
	return
}

// GetByTokenOrIP returns a cached user searched by its session token or IP.
func (l *UserList) GetByTokenOrIP(tokenOrIP string) (u *User, ok bool) {
	u, ok = l.byTokenOrIP[tokenOrIP]
	return
}

// MakeAndAdd creates an *User with the given DBUser, adds it
// to the user list and returns the *User.
func (l *UserList) MakeAndAdd(dbUser *DBUser, tokenOrIP string) (*User, error) {
	u := MakeUser(dbUser)

	_, idOk := l.byID[u.ID]
	_, tokenOk := l.byTokenOrIP[tokenOrIP]
	if idOk || tokenOk {
		return nil, fmt.Errorf("cannot add user already in user list")
	}

	l.byID[u.ID] = u
	l.byTokenOrIP[tokenOrIP] = u
	return u, nil
}

// MakeUserList creates a new UserList.
func MakeUserList() *UserList {
	return &UserList{
		byID:        make(map[uint]*User),
		byTokenOrIP: make(map[string]*User),
	}
}

func IsUserChatBanned(user *User) bool {
	return user.IsPermanentlyChatBanned //TODO(link)|| user.ChatBanExpiry > time.Now()
}
