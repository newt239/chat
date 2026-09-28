package entity

import "testing"

func TestPermissionMatrixAllows(t *testing.T) {
	matrix := DefaultPermissionMatrix().Apply([]PermissionOverride{
		{Role: WorkspaceRoleMember, Permission: PermissionInviteMembers, Allowed: true},
		{Role: WorkspaceRoleAdmin, Permission: PermissionPinMessages, Allowed: false},
	})

	tests := []struct {
		name       string
		role       WorkspaceRole
		permission Permission
		want       bool
	}{
		{name: "オーナーは設定に関係なく許可される", role: WorkspaceRoleOwner, permission: PermissionDeleteOthersMessages, want: true},
		{name: "既定で管理者はすべて許可される", role: WorkspaceRoleAdmin, permission: PermissionDeleteOthersMessages, want: true},
		{name: "上書きで管理者の権限を外せる", role: WorkspaceRoleAdmin, permission: PermissionPinMessages, want: false},
		{name: "既定でメンバーはチャンネルを作れる", role: WorkspaceRoleMember, permission: PermissionCreatePublicChannel, want: true},
		{name: "既定でメンバーは他人のメッセージを削除できない", role: WorkspaceRoleMember, permission: PermissionDeleteOthersMessages, want: false},
		{name: "上書きでメンバーに招待を許可できる", role: WorkspaceRoleMember, permission: PermissionInviteMembers, want: true},
		{name: "既定でゲストは何もできない", role: WorkspaceRoleGuest, permission: PermissionPinMessages, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := matrix.Allows(tt.role, tt.permission); got != tt.want {
				t.Errorf("got=%v want=%v", got, tt.want)
			}
		})
	}
}
