package entity

type Permission string

const (
	PermissionCreatePublicChannel  Permission = "create_public_channel"
	PermissionCreatePrivateChannel Permission = "create_private_channel"
	PermissionInviteMembers        Permission = "invite_members"
	PermissionEditUserGroups       Permission = "edit_user_groups"
	PermissionEditChannelLinks     Permission = "edit_channel_links"
	PermissionPinMessages          Permission = "pin_messages"
	PermissionDeleteOthersMessages Permission = "delete_others_messages"
	PermissionExportData           Permission = "export_data"
	PermissionAddExternalApps      Permission = "add_external_apps"
)

// AllPermissions は権限の一覧を表示順に並べたものです
var AllPermissions = []Permission{
	PermissionCreatePublicChannel,
	PermissionCreatePrivateChannel,
	PermissionInviteMembers,
	PermissionEditUserGroups,
	PermissionEditChannelLinks,
	PermissionPinMessages,
	PermissionDeleteOthersMessages,
	PermissionExportData,
	PermissionAddExternalApps,
}

// ConfigurableRoles は権限を設定できるロールです。オーナーは常にすべての操作ができます
var ConfigurableRoles = []WorkspaceRole{WorkspaceRoleAdmin, WorkspaceRoleMember, WorkspaceRoleGuest}

func (p Permission) IsValid() bool {
	for _, known := range AllPermissions {
		if p == known {
			return true
		}
	}
	return false
}

func IsConfigurableRole(role WorkspaceRole) bool {
	for _, r := range ConfigurableRoles {
		if r == role {
			return true
		}
	}
	return false
}

// PermissionMatrix はロールごとに許可された操作を表します
type PermissionMatrix map[WorkspaceRole]map[Permission]bool

// DefaultPermissionMatrix は保存された設定がない場合の権限です
func DefaultPermissionMatrix() PermissionMatrix {
	admin := make(map[Permission]bool, len(AllPermissions))
	for _, p := range AllPermissions {
		admin[p] = true
	}
	return PermissionMatrix{
		WorkspaceRoleAdmin: admin,
		WorkspaceRoleMember: {
			PermissionCreatePublicChannel:  true,
			PermissionCreatePrivateChannel: true,
			PermissionEditUserGroups:       true,
			PermissionEditChannelLinks:     true,
			PermissionPinMessages:          true,
		},
		WorkspaceRoleGuest: {},
	}
}

// PermissionOverride は既定値から変更された権限です
type PermissionOverride struct {
	Role       WorkspaceRole
	Permission Permission
	Allowed    bool
}

func (m PermissionMatrix) Apply(overrides []PermissionOverride) PermissionMatrix {
	for _, o := range overrides {
		if m[o.Role] == nil {
			m[o.Role] = map[Permission]bool{}
		}
		m[o.Role][o.Permission] = o.Allowed
	}
	return m
}

func (m PermissionMatrix) Allows(role WorkspaceRole, p Permission) bool {
	if role == WorkspaceRoleOwner {
		return true
	}
	return m[role][p]
}
