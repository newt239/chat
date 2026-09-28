package channel

import (
	"context"
	"errors"
	"slices"
	"testing"

	"github.com/newt239/chat/internal/domain/entity"
	domerr "github.com/newt239/chat/internal/domain/errors"
	domainrepository "github.com/newt239/chat/internal/domain/repository"
	domainservice "github.com/newt239/chat/internal/domain/service"
	"github.com/newt239/chat/internal/usecase/audit/audittest"
)

const (
	workspaceID = "acme"
	ownerID     = "00000000-0000-0000-0000-000000000001"
	memberID    = "00000000-0000-0000-0000-000000000002"
	guestID     = "00000000-0000-0000-0000-000000000003"
	creatorID   = "00000000-0000-0000-0000-000000000004"
)

type stubWorkspaceRepo struct {
	domainrepository.WorkspaceRepository
}

func (stubWorkspaceRepo) FindByID(_ context.Context, id string) (*entity.Workspace, error) {
	return &entity.Workspace{ID: id}, nil
}

func (stubWorkspaceRepo) FindMember(_ context.Context, _ string, userID string) (*entity.WorkspaceMember, error) {
	roles := map[string]entity.WorkspaceRole{
		ownerID:   entity.WorkspaceRoleOwner,
		memberID:  entity.WorkspaceRoleMember,
		guestID:   entity.WorkspaceRoleGuest,
		creatorID: entity.WorkspaceRoleMember,
	}
	role, ok := roles[userID]
	if !ok {
		return nil, nil
	}
	return &entity.WorkspaceMember{UserID: userID, Role: role}, nil
}

type stubChannelRepo struct {
	domainrepository.ChannelRepository
	channel *entity.Channel
	created int
	updated int
}

func (r *stubChannelRepo) FindByID(context.Context, string) (*entity.Channel, error) {
	return r.channel, nil
}

func (r *stubChannelRepo) Create(context.Context, *entity.Channel) error {
	r.created++
	return nil
}

func (r *stubChannelRepo) Update(context.Context, *entity.Channel) error {
	r.updated++
	return nil
}

type stubChannelMemberRepo struct {
	domainrepository.ChannelMemberRepository
}

func (stubChannelMemberRepo) AddMember(context.Context, *entity.ChannelMember) error {
	return nil
}

type stubPermissionRepo struct {
	domainrepository.PermissionRepository
	overrides []entity.PermissionOverride
}

func (r stubPermissionRepo) FindOverrides(context.Context, string) ([]entity.PermissionOverride, error) {
	return r.overrides, nil
}

type stubTxManager struct{}

func (stubTxManager) Do(ctx context.Context, fn func(context.Context) error) error {
	return fn(ctx)
}

func newInteractor(channel *entity.Channel, overrides ...entity.PermissionOverride) (ChannelUseCase, *stubChannelRepo, *audittest.Recorder) {
	channelRepo := &stubChannelRepo{channel: channel}
	recorder := &audittest.Recorder{}
	permissionSvc := domainservice.NewPermissionService(stubWorkspaceRepo{}, stubPermissionRepo{overrides: overrides})
	uc := NewChannelInteractor(channelRepo, stubChannelMemberRepo{}, stubWorkspaceRepo{}, nil, stubTxManager{}, nil, nil, permissionSvc, recorder)
	return uc, channelRepo, recorder
}

func TestCreateChannelPermission(t *testing.T) {
	denyPrivate := entity.PermissionOverride{Role: entity.WorkspaceRoleMember, Permission: entity.PermissionCreatePrivateChannel, Allowed: false}
	tests := []struct {
		name      string
		userID    string
		isPrivate bool
		overrides []entity.PermissionOverride
		wantErr   error
	}{
		{name: "既定ではメンバーは公開チャンネルを作れる", userID: memberID},
		{name: "既定ではゲストはチャンネルを作れない", userID: guestID, wantErr: domerr.ErrUnauthorized},
		{name: "非公開チャンネルの作成を禁止するとメンバーは作れない", userID: memberID, isPrivate: true, overrides: []entity.PermissionOverride{denyPrivate}, wantErr: domerr.ErrUnauthorized},
		{name: "非公開チャンネルを禁止しても公開チャンネルは作れる", userID: memberID, overrides: []entity.PermissionOverride{denyPrivate}},
		{name: "オーナーは設定に関係なく作れる", userID: ownerID, isPrivate: true, overrides: []entity.PermissionOverride{{Role: entity.WorkspaceRoleOwner, Permission: entity.PermissionCreatePrivateChannel, Allowed: false}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc, repo, recorder := newInteractor(nil, tt.overrides...)
			_, err := uc.CreateChannel(context.Background(), CreateChannelInput{WorkspaceID: workspaceID, UserID: tt.userID, Name: "general", IsPrivate: tt.isPrivate})
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("エラーが期待と異なります: got=%v want=%v", err, tt.wantErr)
			}
			wantCreated := 0
			var wantActions []entity.AuditAction
			if tt.wantErr == nil {
				wantCreated = 1
				wantActions = []entity.AuditAction{entity.AuditActionChannelCreated}
			}
			if repo.created != wantCreated || !slices.Equal(recorder.Actions(), wantActions) {
				t.Errorf("作成と監査ログが期待と異なります: created=%d actions=%v", repo.created, recorder.Actions())
			}
		})
	}
}

func TestSetArchived(t *testing.T) {
	tests := []struct {
		name        string
		channel     entity.Channel
		userID      string
		wantErr     error
		wantActions []entity.AuditAction
	}{
		{name: "作成者はアーカイブできる", channel: entity.Channel{Type: entity.ChannelTypePublic, CreatedBy: creatorID}, userID: creatorID, wantActions: []entity.AuditAction{entity.AuditActionChannelArchived}},
		{name: "管理者はアーカイブできる", channel: entity.Channel{Type: entity.ChannelTypePublic, CreatedBy: creatorID}, userID: ownerID, wantActions: []entity.AuditAction{entity.AuditActionChannelArchived}},
		{name: "作成者以外のメンバーはアーカイブできない", channel: entity.Channel{Type: entity.ChannelTypePublic, CreatedBy: creatorID}, userID: memberID, wantErr: ErrUnauthorized},
		{name: "DM はアーカイブできない", channel: entity.Channel{Type: entity.ChannelTypeDM, CreatedBy: creatorID}, userID: creatorID, wantErr: ErrCannotArchiveDM},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ch := tt.channel
			ch.WorkspaceID = workspaceID
			uc, _, recorder := newInteractor(&ch)
			out, err := uc.SetArchived(context.Background(), SetArchivedInput{ChannelID: "ch", UserID: tt.userID, Archived: true})
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("エラーが期待と異なります: got=%v want=%v", err, tt.wantErr)
			}
			if tt.wantErr == nil && out.ArchivedAt == nil {
				t.Errorf("アーカイブ日時が設定されていません")
			}
			if !slices.Equal(recorder.Actions(), tt.wantActions) {
				t.Errorf("監査ログが期待と異なります: %v", recorder.Actions())
			}
		})
	}
}

func TestSetArchivedIsIdempotent(t *testing.T) {
	uc, repo, recorder := newInteractor(&entity.Channel{WorkspaceID: workspaceID, Type: entity.ChannelTypePublic, CreatedBy: creatorID})
	if _, err := uc.SetArchived(context.Background(), SetArchivedInput{ChannelID: "ch", UserID: creatorID, Archived: false}); err != nil {
		t.Fatalf("予期しないエラー: %v", err)
	}
	if repo.updated != 0 || len(recorder.Logs) != 0 {
		t.Errorf("状態が変わらない場合は更新も記録もしないはず")
	}
}
