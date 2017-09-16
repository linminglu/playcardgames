package mod

import (
	"fmt"
	mdtime "playcards/model/time"
	pbu "playcards/proto/user"
	"time"
)

type User struct {
	UserID        int32  `gorm:"primary_key"`
	Username      string `reg:"required,min=6,max=32,excludesall= 	"`
	Password      string `reg:"required,min=6,max=32,excludesall= 	"`
	Nickname      string `reg:"required,min=6,max=16,excludesall= 	"`
	Mobile        string `reg:"omitempty,min=6,max=16,excludesall= 	"`
	Email         string `reg:"min=6,max=30,email,excludesall= 	"`
	Avatar        string
	Status        int32
	Channel       string `reg:"omitempty,min=6,max=64,excludesall= 	"`
	Icon          string
	Sex           int32
	Rights        int32
	CreatedAt     *time.Time
	UpdatedAt     *time.Time
	LastLoginAt   *time.Time
	InviteUserID  int32
	MobileUuID    string
	MobileModel   string
	MobileNetWork string
	MobileOs      string
	LastLoginIP   string
	RegIP         string
	OpenID        string
	UnionID       string
	RoomID        int32 `gorm:"-"` // Ignore this field
	Diamond       int64 `gorm:"-"`
}

type UserInfo struct {
	UserID   int32
	Username string
	Nickname string
	Icon     string
	Sex      int32
}

func (u *User) String() string {
	return fmt.Sprintf("[uid: %d, name: %s, rights: %d]", u.UserID,
		u.Username, u.Rights)
}

func (u *User) ToProto() *pbu.User {
	return &pbu.User{
		UserID:        u.UserID,
		Username:      u.Username,
		Password:      u.Password,
		Nickname:      u.Nickname,
		Mobile:        u.Mobile,
		Email:         u.Email,
		Avatar:        u.Avatar,
		Status:        u.Status,
		Channel:       u.Channel,
		Rights:        u.Rights,
		CreatedAt:     mdtime.TimeToProto(u.CreatedAt),
		UpdatedAt:     mdtime.TimeToProto(u.UpdatedAt),
		LastLoginAt:   mdtime.TimeToProto(u.LastLoginAt),
		InviteUserID:  u.InviteUserID,
		MobileUuID:    u.MobileUuID,
		MobileModel:   u.MobileModel,
		MobileNetWork: u.MobileNetWork,
		MobileOs:      u.MobileOs,
		LastLoginIP:   u.LastLoginIP,
		RegIP:         u.RegIP,
		Icon:          u.Icon,
		Sex:           u.Sex,
		OpenID:        u.OpenID,
		UnionID:       u.UnionID,
		Diamond:       u.Diamond,
	}
}
func (u *UserInfo) ToProto() *pbu.UserInfo {
	return &pbu.UserInfo{
		UserID:   u.UserID,
		Username: u.Username,
		Nickname: u.Nickname,
		Icon:     u.Icon,
		Sex:      u.Sex,
	}
}

func UserFromPageRequestProto(u *pbu.PageUserListRequest) *User {
	return &User{
		UserID:   u.UserID,
		Username: u.Username,
		Nickname: u.Nickname,
		Rights:   u.Rights,
		OpenID:   u.OpenID,
		UnionID:  u.UnionID,
	}
}

func UserFromProto(u *pbu.User) *User {
	return &User{
		UserID:        u.UserID,
		Username:      u.Username,
		Password:      u.Password,
		Nickname:      u.Nickname,
		Mobile:        u.Mobile,
		Email:         u.Email,
		Avatar:        u.Avatar,
		Status:        u.Status,
		Channel:       u.Channel,
		Rights:        u.Rights,
		CreatedAt:     mdtime.TimeFromProto(u.CreatedAt),
		UpdatedAt:     mdtime.TimeFromProto(u.UpdatedAt),
		LastLoginAt:   mdtime.TimeFromProto(u.LastLoginAt),
		InviteUserID:  u.InviteUserID,
		MobileUuID:    u.MobileUuID,
		MobileModel:   u.MobileModel,
		MobileNetWork: u.MobileNetWork,
		MobileOs:      u.MobileOs,
		LastLoginIP:   u.LastLoginIP,
		RegIP:         u.RegIP,
		Icon:          u.Icon,
		Sex:           u.Sex,
		OpenID:        u.OpenID,
		UnionID:       u.UnionID,
	}
}

func (u *User) AfterFind() error {
	u.Password = ""
	return nil
}
