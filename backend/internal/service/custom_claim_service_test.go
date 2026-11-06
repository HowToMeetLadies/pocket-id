package service

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/pocket-id/pocket-id/backend/internal/appconfig"
	"github.com/pocket-id/pocket-id/backend/internal/apperror"
	"github.com/pocket-id/pocket-id/backend/internal/model"
	datatype "github.com/pocket-id/pocket-id/backend/internal/model/types"
	testutils "github.com/pocket-id/pocket-id/backend/internal/utils/testing"
)

func TestCustomClaimUpdatesRejectMissingOwner(t *testing.T) {
	service := NewCustomClaimService(testutils.NewDatabaseForTest(t))

	_, err := service.UpdateCustomClaimsForUser(t.Context(), "missing-user", nil)
	require.True(t, apperror.IsCode(err, apperror.CodeUserNotFound))

	_, err = service.UpdateCustomClaimsForUserGroup(t.Context(), "missing-group", nil)
	require.True(t, apperror.IsCode(err, apperror.CodeNotFound))
}

func TestGetCustomClaimsForUserWithUserGroupsUsesOrder(t *testing.T) {
	createdOld := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	createdNew := time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name            string
		order           string
		oldPriority     int
		newPriority     int
		oldFriendlyName string
		newFriendlyName string
		want            string
	}{
		{
			name:            "the newest group wins by default",
			oldFriendlyName: "Zulu",
			newFriendlyName: "Alpha",
			want:            "new",
		},
		{
			name:            "the group name can be the tie-break",
			order:           appconfig.CustomClaimGroupOrderName,
			oldFriendlyName: "Alpha",
			newFriendlyName: "Zulu",
			want:            "old",
		},
		{
			name:            "the oldest group can win",
			order:           appconfig.CustomClaimGroupOrderOldest,
			oldFriendlyName: "Alpha",
			newFriendlyName: "Zulu",
			want:            "old",
		},
		{
			name:            "higher priority wins over the newest date",
			order:           appconfig.CustomClaimGroupOrderNewest,
			oldPriority:     10,
			oldFriendlyName: "Zulu",
			newFriendlyName: "Alpha",
			want:            "old",
		},
		{
			name:            "higher priority wins over the group name",
			order:           appconfig.CustomClaimGroupOrderName,
			oldPriority:     10,
			oldFriendlyName: "Zulu",
			newFriendlyName: "Alpha",
			want:            "old",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			customClaimService := NewCustomClaimService(testutils.NewDatabaseForTest(t))
			db := customClaimService.db

			user := model.User{Username: "claim-user"}
			require.NoError(t, db.Create(&user).Error)

			createClaimGroup(t, db, "group-old", tt.oldFriendlyName, tt.oldPriority, createdOld, "old")
			createClaimGroup(t, db, "group-new", tt.newFriendlyName, tt.newPriority, createdNew, "new")

			require.NoError(t, db.Exec(
				"INSERT INTO user_groups_users (user_id, user_group_id) VALUES (?, ?), (?, ?)",
				user.ID, "group-old", user.ID, "group-new",
			).Error)

			claims, err := customClaimService.GetCustomClaimsForUserWithUserGroupsWithOrder(t.Context(), user.ID, db, tt.order)
			require.NoError(t, err)
			require.Equal(t, tt.want, claimValue(claims, "role"))
		})
	}
}

func TestGetCustomClaimsForUserWithUserGroupsKeepsUserClaims(t *testing.T) {
	customClaimService := NewCustomClaimService(testutils.NewDatabaseForTest(t))
	db := customClaimService.db

	user := model.User{Username: "claim-user"}
	require.NoError(t, db.Create(&user).Error)
	require.NoError(t, db.Create(&model.CustomClaim{
		Key:    "role",
		Value:  "user",
		UserID: &user.ID,
	}).Error)

	now := time.Now()
	createClaimGroup(t, db, "group-a", "Alpha", 10, now, "group-a")
	createClaimGroup(t, db, "group-b", "Beta", 0, now.Add(-time.Hour), "group-b")
	require.NoError(t, db.Exec(
		"INSERT INTO user_groups_users (user_id, user_group_id) VALUES (?, ?), (?, ?)",
		user.ID, "group-a", user.ID, "group-b",
	).Error)

	claims, err := customClaimService.GetCustomClaimsForUserWithUserGroupsWithOrder(t.Context(), user.ID, db, appconfig.CustomClaimGroupOrderNewest)
	require.NoError(t, err)
	require.Len(t, claims, 1)
	require.Equal(t, "user", claimValue(claims, "role"))
}

func createClaimGroup(t *testing.T, db *gorm.DB, id, friendlyName string, priority int, createdAt time.Time, value string) {
	t.Helper()

	group := model.UserGroup{
		Base:          model.Base{ID: id},
		FriendlyName:  friendlyName,
		Name:          id,
		ClaimPriority: priority,
	}
	require.NoError(t, db.Create(&group).Error)
	require.NoError(t, db.
		Model(&model.UserGroup{}).
		Where("id = ?", id).
		UpdateColumn("created_at", datatype.DateTime(createdAt)).
		Error)
	require.NoError(t, db.Create(&model.CustomClaim{
		Key:         "role",
		Value:       value,
		UserGroupID: &id,
	}).Error)
}

func claimValue(claims []model.CustomClaim, key string) string {
	for _, claim := range claims {
		if claim.Key == key {
			return claim.Value
		}
	}
	return ""
}
