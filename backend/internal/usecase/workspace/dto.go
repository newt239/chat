package workspace

import "github.com/newt239/chat/internal/domain/entity"

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

type WorkspaceOutput struct {
	*entity.Workspace
	Role entity.WorkspaceRole
}

type MemberInfo struct {
	*entity.WorkspaceMember
	User *entity.User
	// 取得したユーザーだけに見えるニックネーム
	Nickname *string
}

// NewMemberInfos は users にないメンバーを除きます
func NewMemberInfos(members []*entity.WorkspaceMember, users map[string]*entity.User) []MemberInfo {
	infos := make([]MemberInfo, 0, len(members))
	for _, m := range members {
		if user := users[m.UserID]; user != nil {
			infos = append(infos, MemberInfo{WorkspaceMember: m, User: user})
		}
	}
	return infos
}

type PublicWorkspaceItem struct {
	*entity.Workspace
	MemberCount int
	IsJoined    bool
}
