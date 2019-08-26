package main

import (
	"fmt"
	"net/url"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
)

// Database wraps an SQL database with helper methods.
type Database struct {
	sql    *gorm.DB
	driver string
}

// GetUserByID returns the user with the given ID.
func (db *Database) GetUserByID(id uint) (u *DBUser, err error) {
	u = new(DBUser)
	err = db.sql.First(u, "id = ?", id).Error
	return
}

// GetUserByLogin returns the user with the given login information.
func (db *Database) GetUserByLogin(login UserLogin) (u *DBUser, err error) {
	u = new(DBUser)
	err = db.sql.First(u, "login = ?", login.String()).Error
	return
}

// GetUserByToken returns the user with the given session token.
func (db *Database) GetUserByToken(token string) (u *DBUser, err error) {
	s := new(DBSession)
	if err = db.sql.First(s, "token = ?", token).Error; err != nil {
		return nil, err
	}

	u, err = db.GetUserByID(s.UserID)
	return
}

// CreateUser creates an user with the given name, login, ip and user agent.
func (db *Database) CreateUser(name string, login UserLogin, ip, ua string) (*DBUser, error) {
	user := &DBUser{
		Name:      name,
		Login:     login,
		SignupIP:  ip,
		LastIP:    ip,
		UserAgent: ua,
	}

	var c uint
	db.sql.Model(user).Where("login = ?", user.Login.String()).Count(&c)
	if c > 0 {
		return nil, fmt.Errorf("user with same login (%s) already in database", user.Login.String())
	}
	db.sql.Create(user)

	return user, nil
}

// SaveSessionForUser creates a session in the database with the given user ID and token.
func (db *Database) SaveSessionForUser(uid uint, token string) error {
	return db.sql.Create(&DBSession{
		UserID: uid,
		Token:  token,
	}).Error
}

// PlacePixel inserts a pixel into the database.
func (db *Database) PlacePixel(x, y uint, color byte, placer *User) error {
	tx := db.sql.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	/// Unset IsMostRecent on last pixel
	if oldPixel := new(DBPixel); !tx.First(oldPixel, "x = ? AND y = ? AND most_recent", x, y).RecordNotFound() {
		oldPixel.IsMostRecent = false
		tx.Save(oldPixel)
	}

	/// Save pixel
	pixel := &DBPixel{
		PosX:     x,
		PosY:     y,
		ColorIdx: color,
	}
	if placer != nil {
		pixel.PlacerID = placer.ID
	}
	tx.Save(pixel)

	/// Update user pixel counts
	placer.DBUser.PixelCount++
	placer.DBUser.PixelCountAlltime++
	tx.Save(placer.DBUser)

	return tx.Commit().Error
}

// SetUserCooldownExpiry sets the cooldown expiry timestamp of the user with the given ID.
func (db *Database) SetUserCooldownExpiry(uid uint, ce time.Time) error {
	u := &DBUser{CooldownExpiry: &ce}
	return db.sql.Model(u).Update(u).Where("id = ?", uid).Error
}

// Close closes the internal connection to the database.
func (db *Database) Close() error {
	return db.sql.Close()
}

// MakeDatabase creates and connects to the database.
func MakeDatabase(driver, user, pass, uri string) (*Database, error) {
	// https://github.com/pxlsspace/Pxls/blob/master/src/main/java/space/pxls/data/Database.java#L49
	pURI, err := url.Parse(uri)
	if err != nil {
		return nil, err
	}

	connConf := mysql.Config{
		Net:                  "tcp",
		Addr:                 pURI.Host,
		DBName:               pURI.Path[1:],
		User:                 user,
		Passwd:               pass,
		MultiStatements:      true,
		ParseTime:            true,
		AllowNativePasswords: true,
	}
	conn, err := gorm.Open(driver, connConf.FormatDSN())
	if err != nil {
		return nil, err
	}

	// Generate tables and migrate them when a difference with the models is detected.
	conn.AutoMigrate(&DBPixel{}, &DBUser{}, &DBSession{})

	return &Database{
		sql:    conn,
		driver: driver,
	}, nil
}

// DBPixel represents a pixel as stored in the database
type DBPixel struct {
	ID       uint       `gorm:"not null; primary_key; auto_increment"`
	PosX     uint       `gorm:"column:x; not null; index:pos"`
	PosY     uint       `gorm:"column:y; not null; index:pos"`
	PlacerID uint       `gorm:"column:who"`
	ColorIdx byte       `gorm:"column:color; not null"`
	Time     *time.Time `gorm:"type:timestamp; not null; default:now(6)"`

	// TODO(netux): implement these
	// Secondary ID is the previous pixel's ID.
	// If the pixel was rollbacked, this is the ID that was changed from for rollback action,
	// is NULL if there's no previous or it was undo of rollback
	// SecondaryID uint
	// IsModAction bool `gorm:"column:mod_action; not null; default:false"`
	// RollbackAction bool `gorm:"not null; default:false"`
	// Undone         byte `gorm:"not null; default:0"`
	// UndoAction     bool `gorm:"not null; default:false"`

	IsMostRecent bool `gorm:"column:most_recent; not null; default:true; index:most_recent"`
}

// TableName returns the name of the pixels table.
func (*DBPixel) TableName() string {
	return "pixels"
}

// DBUser represents an user as stored in the database
type DBUser struct {
	ID                uint     `gorm:"not null; primary_key; auto_increment"`
	Name              string   `gorm:"column:username; type:varchar(32); not null"`
	Role              UserRole `gorm:"type:varchar(16); not null; default:'USER'"`
	PixelCount        uint64   `gorm:"not null; default:0"`
	PixelCountAlltime uint64   `gorm:"not null; default:0"`

	RawLogin string    `gorm:"column:login; type:varchar(64); not null"`
	Login    UserLogin `gorm:"-"`

	SignupTime *time.Time `gorm:"type:timestamp; not null; default:now(6)"`
	SignupIP   string     `gorm:"type:varbinary(16)"`
	LastIP     string     `gorm:"type:varbinary(16)"`
	UserAgent  string     `gorm:"type:varchar(512); not null; default:''"`

	Stacked        int        `gorm:"default:0"`
	CooldownExpiry *time.Time `gorm:"type:timestamp"`

	BanExpiry               *time.Time `gorm:"type:timestamp"`
	BanReason               string     `gorm:"type:varchar(512); not null; default:''"`
	ChatBanExpiry           *time.Time `gorm:"type:timestamp; default:now()"`
	ChatBanReason           string     `gorm:"type:text"`
	IsPermanentlyChatBanned bool       `gorm:"column:perma_chat_banned; default:false"`
	IsRenameRequested       bool       `gorm:"not null; default:false"`
}

// BeforeSave updates the raw login details before saving to the database.
func (u *DBUser) BeforeSave() error {
	u.RawLogin = u.Login.String()
	return nil
}

// AfterFind updates the login details after fetching the user from the database.
func (u *DBUser) AfterFind() error {
	u.Login = ParseUserLogin(u.RawLogin)
	return nil
}

// TableName returns the name of the users table.
func (*DBUser) TableName() string {
	return "users"
}

// DBSession represents a session (user + token) as stored in the database.
type DBSession struct {
	ID     uint      `gorm:"not null; primary_key; auto_increment"`
	UserID uint      `gorm:"column:who; not null"`
	Token  string    `gorm:"type:varchar(60); not null; unique"`
	Time   time.Time `gorm:"type:timestamp; default:now()"`
}

// BeforeUpdate updates the session's time locally before committing to the database.
func (s *DBSession) BeforeUpdate() error {
	// TODO(netux): check if this behaves like "ON UPDATE CURRENT_TIMESTAMP"
	s.Time = time.Now()
	return nil
}

// TableName returns the name of the sessions table.
func (*DBSession) TableName() string {
	return "sessions"
}
