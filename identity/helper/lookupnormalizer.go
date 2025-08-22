package helper

import (
	"context"
	"errors"
	"net/mail"
	"strings"

	"p9e.in/ugcl/identity/models"

	"github.com/nyaruka/phonenumbers"
)

type LookupNormalizer interface {
	// Name normalizer
	Name(ctx context.Context, name string) (string, error)
	// Email normalizer
	Email(ctx context.Context, email string) (string, error)
	// Phone normalizer
	Phone(ctx context.Context, phone string) (string, error)
	Normalize(context.Context, *models.User) error
}

type lookupNormalizer struct {
}

func NewLookupNormalizer() LookupNormalizer {
	return lookupNormalizer{}
}
func (l lookupNormalizer) Name(ctx context.Context, name string) (string, error) {
	if name == "" {
		return "", errors.New("invalid username")
	}
	if _, err := l.Email(ctx, name); err == nil {
		return "", errors.New("invalid username")
	}
	if _, err := l.Phone(ctx, name); err == nil {
		return "", errors.New("invalid username")
	}
	return strings.ToUpper(name), nil
}

func (l lookupNormalizer) Email(ctx context.Context, email string) (string, error) {
	if email == "" {
		return "", errors.New("invalid email")
	}
	if _, err := mail.ParseAddress(email); err != nil {
		return "", errors.New("invalid email")
	}
	return strings.ToUpper(email), nil
}

func (l lookupNormalizer) Phone(ctx context.Context, phone string) (string, error) {
	if phone == "" {
		return "", errors.New("invalid phone")
	}
	num, err := phonenumbers.Parse(phone, "US")
	if err != nil {
		return "", errors.New("invalid phone")
	}
	if ok := phonenumbers.IsValidNumber(num); !ok {
		return "", errors.New("invalid phone")
	}
	formattedNum := phonenumbers.Format(num, phonenumbers.E164)
	return formattedNum, err
}

func (l lookupNormalizer) Normalize(ctx context.Context, u *models.User) error {
	//normalize
	if u.Username != nil {
		n, err := l.Name(ctx, *u.Username)
		if err != nil {
			return err
		}
		u.NormalizedUsername = &n
	}
	if u.Email != nil {
		e, err := l.Email(ctx, *u.Email)
		if err != nil {
			return err
		}
		u.NormalizedEmail = &e
	}
	if u.Phone != nil {
		phone, err := l.Phone(ctx, *u.Phone)
		if err != nil {
			return err
		}
		u.Phone = &phone
	}
	// t, _ := saas.FromCurrentTenant(ctx)
	// if len(t.GetId()) > 0 {
	// 	ti := t.GetId()
	// 	u.CreatedTenant = &ti
	// }
	return nil
}
