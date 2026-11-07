package dto

import (
	"errors"

	"github.com/gin-gonic/gin/binding"
	datatype "github.com/pocket-id/pocket-id/backend/internal/model/types"
)

type UserGroupDto struct {
	ID                 string                  `json:"id"`
	FriendlyName       string                  `json:"friendlyName"`
	Name               string                  `json:"name"`
	ClaimPriority      int                     `json:"claimPriority"`
	CustomClaims       []CustomClaimDto        `json:"customClaims"`
	LdapID             *string                 `json:"ldapId"`
	CreatedAt          datatype.DateTime       `json:"createdAt"`
	Users              []UserDto               `json:"users"`
	AllowedOidcClients []OidcClientMetaDataDto `json:"allowedOidcClients"`
}

type UserGroupMinimalDto struct {
	ID            string            `json:"id"`
	FriendlyName  string            `json:"friendlyName"`
	Name          string            `json:"name"`
	ClaimPriority int               `json:"claimPriority"`
	CustomClaims  []CustomClaimDto  `json:"customClaims"`
	UserCount     int64             `json:"userCount"`
	LdapID        *string           `json:"ldapId"`
	CreatedAt     datatype.DateTime `json:"createdAt"`
}

type UserGroupUpdateAllowedOidcClientsDto struct {
	OidcClientIDs []string `json:"oidcClientIds" binding:"required"`
}

type UserGroupCreateDto struct {
	FriendlyName  string `json:"friendlyName" binding:"required,min=2,max=50" unorm:"nfc"`
	Name          string `json:"name" binding:"required,min=2,max=255" unorm:"nfc"`
	ClaimPriority *int   `json:"claimPriority" binding:"omitempty,gte=0,lte=1000"`
	LdapID        string `json:"-"`
}

func (g UserGroupCreateDto) Validate() error {
	e, ok := binding.Validator.Engine().(interface {
		Struct(s any) error
	})
	if !ok {
		return errors.New("validator does not implement the expected interface")
	}

	return e.Struct(g)
}

type UserGroupUpdateUsersDto struct {
	UserIDs []string `json:"userIds" binding:"required"`
}
