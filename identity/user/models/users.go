package models

import (
	"sync"
	"time"

	"p9e.in/ugcl/packages/models"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"google.golang.org/protobuf/types/known/wrapperspb"
	pb "p9e.in/ugcl/identity/user/api/v2/user"
)

type Gender int

const (
	GenderMale Gender = iota
	GenderFemale
	GenderOther
)

type User struct {
	Uuid               uuid.UUID      `db:"uuid"`
	Id                 int32          `db:"id"`
	Username           *string        `db:"username"`
	NormalizedUsername *string        `db:"normalized_username"`
	Fullname           *string        `db:"fullname"`
	Phone              *string        `db:"phone"`
	PhoneConfirmed     bool           `db:"phone_confirmed"`
	Email              *string        `db:"email"`
	NormalizedEmail    *string        `db:"normalized_email"`
	EmailConfirmed     bool           `db:"email_confirmed"`
	Password           string         `db:"password"`
	Gender             Gender         `db:"gender"`
	TenantIds          pq.StringArray `db:"tenant_ids"`
	TenantRoles        pq.StringArray `db:"tenant_roles"`
	Avatar             []byte         `db:"avatar"`
	TwoFactorEnabled   bool           `db:"two_factor_enabled"`
	PasswordHash       string         `db:"password_hash" validate:"required"`
	Salt               []byte         `db:"salt"`
	TwoFactorSecret    *string        `db:"two_factor_secret"`
	IsActive           bool           `db:"is_active"`
	CreatedAt          time.Time      `db:"created_at"`
	UpdatedAt          time.Time      `db:"updated_at"`
	DeletedAt          *time.Time     `db:"deleted_at"`
}

type UserSearchCriteria struct {
	*models.SearchCriteria
	UsernameSearch *string `json:"username_search,omitempty"`
	PhoneSearch    *string `json:"phone_search,omitempty"`
	EmailSearch    *string `json:"email_search,omitempty"`
	FullnameSearch *string `json:"fullname_search,omitempty"`
}

type UserIdentifier struct {
	Uuid string
	Id   int32
}

var (
	userProtoPool = sync.Pool{
		New: func() interface{} {
			return &pb.User{
				Uuid:     wrapperspb.String(""),
				Username: wrapperspb.String(""),
				Fullname: wrapperspb.String(""),
				Phone:    wrapperspb.String(""),
				Email:    wrapperspb.String(""),
			}
		},
	}

	userDBPool = sync.Pool{
		New: func() interface{} {
			return &User{
				TenantIds:   make([]string, 0, 5),
				TenantRoles: make([]string, 0, 5),
				Avatar:      make([]byte, 0, 100),
				Salt:        make([]byte, 0, 16),
			}
		},
	}
)

func UserToProto(userDB *User) *pb.User {
	user := userProtoPool.Get().(*pb.User)
	user.Reset()

	user.Id = userDB.Id
	user.Uuid = wrapperspb.String(userDB.Uuid.String())

	if userDB.Username != nil {
		user.Username = wrapperspb.String(*userDB.Username)
	}
	if userDB.Fullname != nil {
		user.Fullname = wrapperspb.String(*userDB.Fullname)
	}
	if userDB.Phone != nil {
		user.Phone = wrapperspb.String(*userDB.Phone)
	}
	if userDB.Email != nil {
		user.Email = wrapperspb.String(*userDB.Email)
	}

	user.Gender = pb.Gender(userDB.Gender)
	user.Password = userDB.Password
	user.TwoFactorEnabled = userDB.TwoFactorEnabled

	if userDB.TwoFactorSecret != nil {
		user.TwoFactorSecret = wrapperspb.String(*userDB.TwoFactorSecret)
	}

	user.IsActive = userDB.IsActive

	return user
}

func ReleaseProtoUser(user *pb.User) {
	if user != nil {
		userProtoPool.Put(user)
	}
}

func GetDBUser() *User {
	return userDBPool.Get().(*User)
}

func ReleaseDBUser(user *User) {
	if user == nil {
		return
	}
	user.TenantIds = user.TenantIds[:0]
	user.TenantRoles = user.TenantRoles[:0]
	user.Avatar = user.Avatar[:0]
	user.Salt = user.Salt[:0]

	user.Uuid = uuid.UUID{}
	user.Id = 0
	user.Username = nil
	user.NormalizedUsername = nil
	user.Fullname = nil
	user.Phone = nil
	user.PhoneConfirmed = false
	user.Email = nil
	user.NormalizedEmail = nil
	user.EmailConfirmed = false
	user.Password = ""
	user.Gender = 0
	user.TwoFactorEnabled = false
	user.PasswordHash = ""
	user.TwoFactorSecret = nil
	user.IsActive = false
	user.CreatedAt = time.Time{}
	user.UpdatedAt = time.Time{}
	user.DeletedAt = nil

	userDBPool.Put(user)
}

func UserToDB(userProto *pb.User) *User {
	if userProto == nil {
		return nil
	}

	user := userDBPool.Get().(*User)
	*user = User{
		TenantIds:   user.TenantIds[:0],
		TenantRoles: user.TenantRoles[:0],
		Avatar:      user.Avatar[:0],
		Salt:        make([]byte, 0, 16),
	}

	user.Id = userProto.Id

	if userProto.Uuid != nil {
		parsedUUID, err := uuid.Parse(userProto.Uuid.Value)
		if err == nil {
			user.Uuid = parsedUUID
		}
	}

	if userProto.Username != nil {
		user.Username = &userProto.Username.Value
	}

	if userProto.Fullname != nil {
		user.Fullname = &userProto.Fullname.Value
	}

	if userProto.Phone != nil {
		user.Phone = &userProto.Phone.Value
	}

	if userProto.Email != nil {
		user.Email = &userProto.Email.Value
	}

	user.Gender = Gender(userProto.Gender.Number())
	user.Password = userProto.Password

	if len(userProto.Avatar) > 0 {
		user.Avatar = make([]byte, len(userProto.Avatar))
		copy(user.Avatar, userProto.Avatar)
	}

	user.TwoFactorEnabled = userProto.TwoFactorEnabled

	if userProto.TwoFactorSecret != nil {
		user.TwoFactorSecret = &userProto.TwoFactorSecret.Value
	}

	user.IsActive = userProto.IsActive
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()

	return user
}

func CreateUserRequestToDB(req *pb.CreateUserRequest) *User {
	if req == nil || req.User == nil {
		return nil
	}
	user := UserToDB(req.User)
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()
	return user
}

func UpdateUserRequestToDB(req *pb.UpdateUserRequest) *User {
	if req == nil || req.User == nil {
		return nil
	}
	user := UserToDB(req.User)
	user.UpdatedAt = time.Now()
	return user
}
