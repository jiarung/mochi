package utils

import (
	"fmt"
	"net/http"
	"time"

	"github.com/cobinhood/gorm"
	"github.com/satori/go.uuid"

	"github.com/cobinhood/cobinhood-backend/cache"
	cachehelper "github.com/cobinhood/cobinhood-backend/cache/helper"
	"github.com/cobinhood/cobinhood-backend/cache/keys"
	"github.com/cobinhood/cobinhood-backend/models/exchange"
	"github.com/cobinhood/cobinhood-backend/models/exchange/helper"
	"github.com/cobinhood/cobinhood-backend/types"
)

// ReviewMotionRequest is the request of vote action.
type ReviewMotionRequest struct {
	Action  types.AuditCommitteeVoteType `json:"action" binding:"required,eq=audit_committee_vote_type_approve|eq=audit_committee_vote_type_reject"`
	Comment string                       `json:"comment"`
}

// RequestDescription returns HTTP request description.
func RequestDescription(req *http.Request) string {
	return fmt.Sprintf("[%s] %s", req.Method, req.URL.Path)
}

// CreateAuditLog creates audit log.
func CreateAuditLog(tx *gorm.DB, performerID uuid.UUID,
	performerScopes []types.Scope, performerIP string,
	action types.AuditLogAction, clientID uuid.UUID, desc string) error {
	auditLog := exchange.AuditLog{
		Action:          action,
		PerformerID:     performerID,
		PerformerScopes: types.ScopeSlice(performerScopes),
		PerformerIP:     performerIP,
		ClientID:        &clientID,
		Description:     &desc,
	}
	return tx.Create(&auditLog).Error
}

// RefreshUserScopes refreshes user scopes.
func RefreshUserScopes(tx *gorm.DB, redis *cache.Redis, userID uuid.UUID) error {
	var (
		roleHelper        = helper.NewRoleHelper(tx)
		accessTokenHelper = helper.NewAccessTokenHelper(tx)
		ids               []uuid.UUID
	)
	roles, err := roleHelper.GetUserRoles(userID, true)
	if err != nil {
		return err
	}
	tokens, err := accessTokenHelper.GetActiveTokensByUserID(userID, true)
	if err != nil {
		return err
	}
	ids = make([]uuid.UUID, len(tokens))
	for i := range tokens {
		ids[i] = tokens[i].ID
	}
	if err := accessTokenHelper.UpdateTokensByIDs(ids,
		map[string]interface{}{
			"authorization_roles": types.RoleSlice(roles),
		}); err != nil {
		return err
	}
	for _, token := range tokens {
		payload := &cachehelper.AccessTokenPayload{
			Roles: roles,
		}
		if token.IPAddress != nil {
			payload.IP = *token.IPAddress
		}
		// Set cache.
		if token.ExpireAt == nil {
			if err := payload.Set(token.ID.String()); err != nil {
				redis.Delete(keys.GetAccessTokenCacheKey(token.ID.String()))
			}
		} else {
			remainingSec := int(token.ExpireAt.Unix() - time.Now().Unix())
			if err := payload.Set(token.ID.String(), remainingSec); err != nil {
				redis.Delete(keys.GetAccessTokenCacheKey(token.ID.String()))
			}
		}
	}
	return nil
}

// RefreshUserScopesByEmail refreshes user scopes.
func RefreshUserScopesByEmail(tx *gorm.DB, redis *cache.Redis, email string) error {
	var user exchange.User
	if err := tx.Model(&exchange.User{}).
		Where("email = ?", email).
		Find(&user).Error; err != nil {
		return err
	}
	return RefreshUserScopes(tx, redis, user.ID)
}
