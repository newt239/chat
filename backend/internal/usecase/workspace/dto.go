package workspace

import (
	"time"

	"github.com/newt239/chat/internal/domain/entity"
)

type CreateWorkspaceInput struct {
	ID          string
	Name        string
	Description *string
	IconURL     *string
	IsPublic    bool
	CreatedBy   string
}

type UpdateWorkspaceInput struct {
	ID                 string
	Name               *string
	Description        *string
	IconURL            *string
	IsPublic           *bool
	SignupEnabled      *bool
	EmailSignupEnabled *bool
	UserID             string
}

// MemberInput は OperatorID のユーザーが UserID のメンバーのロールを変えたり外したりします。Role は削除では使わない
type MemberInput struct {
	WorkspaceID string
	UserID      string
	OperatorID  string
	Role        entity.WorkspaceRole
}

type WorkspaceOutput struct {
	*entity.Workspace
	Role entity.WorkspaceRole
}

type MemberInfo struct {
	UserID      string
	Email       string
	DisplayName string
	AvatarURL   *string
	Bio         *string
	Role        entity.WorkspaceRole
	JoinedAt    time.Time
	SuspendedAt *time.Time
	// 取得したユーザーだけに見えるニックネーム
	Nickname *string
	Timezone string
	Links    []string
}

// NewMemberInfos は users にないメンバーをユーザー情報なしで返します
func NewMemberInfos(members []*entity.WorkspaceMember, users map[string]*entity.User) []MemberInfo {
	infos := make([]MemberInfo, 0, len(members))
	for _, m := range members {
		info := MemberInfo{UserID: m.UserID, Role: m.Role, JoinedAt: m.JoinedAt, SuspendedAt: m.SuspendedAt}
		if user := users[m.UserID]; user != nil {
			info.Email = user.Email
			info.DisplayName = user.DisplayName
			info.AvatarURL = user.AvatarURL
			info.Bio = user.Bio
			info.Timezone = user.Preferences.Timezone
			info.Links = user.Links
		}
		infos = append(infos, info)
	}
	return infos
}

type PublicWorkspaceItem struct {
	*entity.Workspace
	MemberCount int
	IsJoined    bool
}
