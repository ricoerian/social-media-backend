package models

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
    ID           uint           `gorm:"primaryKey"`
    Fullname     string         `gorm:"type:varchar(255)"`
    Username     string         `gorm:"type:varchar(100);uniqueIndex"`
    Email        string         `gorm:"type:varchar(100);uniqueIndex"`
    Password     string         `gorm:"type:varchar(255)"`
    PhotoProfile string         `gorm:"type:varchar(255)"`
    JenisKelamin string         `gorm:"type:varchar(50)"`
    TanggalLahir *time.Time     `gorm:"type:date"`      
    // Relasi many-to-many (follow)
    Followers    []*User        `gorm:"many2many:follows;joinForeignKey:FollowingID;JoinReferences:FollowerID"`
    Following    []*User        `gorm:"many2many:follows;joinForeignKey:FollowerID;JoinReferences:FollowingID"`
    Feeds        []Feed         
    Comments     []Comment      
    Chatrooms    []Chatroom     `gorm:"many2many:chatroom_users;"`
    Messages     []Message      
    CreatedAt    time.Time      
    UpdatedAt    time.Time      
    DeletedAt    gorm.DeletedAt `gorm:"index"`
}

type Follow struct {
	FollowerID  uint      `gorm:"primaryKey"`
	FollowingID uint      `gorm:"primaryKey"`
	CreatedAt   time.Time 
	DeletedAt   gorm.DeletedAt `gorm:"index"`
}

type Feed struct {
	ID        uint           `gorm:"primaryKey"`
	Feed      string         
	File      string         
	UserID    uint           
	User      User           
	Reactions []Reaction
	Comments  []Comment      
	CreatedAt time.Time      
	UpdatedAt time.Time      
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

type Comment struct {
	ID        uint           `gorm:"primaryKey"`
	Comment   string         
	File      string         
	FeedID    uint           
	Feed      Feed           
	UserID    uint           
	User      User           
	CreatedAt time.Time      
	UpdatedAt time.Time      
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

type Reaction struct {
	ID        uint      `gorm:"primaryKey"`
	FeedID    uint      
	Feed      Feed      
	UserID    uint      
	User      User      
	Reaction  string    // "like" atau "dislike"
	CreatedAt time.Time
}

type Chatroom struct {
	ID        uint           `gorm:"primaryKey"`
	Name      string         
	OwnerID   uint           // ID user yang membuat group chat (jika group)
	Users     []User         `gorm:"many2many:chatroom_users;"`
	Messages  []Message      
	IsGroup   bool           
	CreatedAt time.Time      
	UpdatedAt time.Time      
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

type Message struct {
	ID          uint           `gorm:"primaryKey"`
	Message     string         
	File        string         
	ChatroomID  uint           
	Chatroom    Chatroom       
	UserID      uint           
	User        User           
	CreatedAt   time.Time      
	UpdatedAt   time.Time      
	DeletedAt   gorm.DeletedAt `gorm:"index"`
};