package message

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/newt239/chat/internal/domain/entity"
	domerr "github.com/newt239/chat/internal/domain/errors"
	domainrepository "github.com/newt239/chat/internal/domain/repository"
	"github.com/newt239/chat/internal/domain/service"
)

type builderMessageRepo struct {
	domainrepository.MessageRepository
	messages  map[string]*entity.Message
	reactions map[string][]*entity.MessageReaction
}

func (r *builderMessageRepo) FindByIDs(_ context.Context, ids []string) ([]*entity.Message, error) {
	found := make([]*entity.Message, 0, len(ids))
	for _, id := range ids {
		if msg := r.messages[id]; msg != nil {
			found = append(found, msg)
		}
	}
	return found, nil
}

func (r *builderMessageRepo) FindReactionsByMessageIDs(_ context.Context, _ []string) (map[string][]*entity.MessageReaction, error) {
	return r.reactions, nil
}

type builderUserRepo struct {
	domainrepository.UserRepository
	users []*entity.User
}

func (r *builderUserRepo) FindByIDs(_ context.Context, _ []string) (map[string]*entity.User, error) {
	users := make(map[string]*entity.User, len(r.users))
	for _, u := range r.users {
		users[u.ID] = u
	}
	return users, nil
}

type builderUserMentionRepo struct {
	domainrepository.MessageUserMentionRepository
}

func (r *builderUserMentionRepo) FindByMessageIDs(_ context.Context, _ []string) ([]*entity.MessageUserMention, error) {
	return nil, nil
}

type builderGroupMentionRepo struct {
	domainrepository.MessageGroupMentionRepository
}

func (r *builderGroupMentionRepo) FindByMessageIDs(_ context.Context, _ []string) ([]*entity.MessageGroupMention, error) {
	return nil, nil
}

type builderLinkRepo struct {
	domainrepository.MessageLinkRepository
	links []*entity.MessageLink
}

func (r *builderLinkRepo) FindByMessageIDs(_ context.Context, _ []string) ([]*entity.MessageLink, error) {
	return r.links, nil
}

type builderAttachmentRepo struct {
	domainrepository.AttachmentRepository
	attachments map[string][]*entity.Attachment
}

func (r *builderAttachmentRepo) FindByMessageIDs(_ context.Context, _ []string) (map[string][]*entity.Attachment, error) {
	return r.attachments, nil
}

type builderPinRepo struct {
	domainrepository.PinRepository
	pins map[string]*entity.MessagePin
}

func (r *builderPinRepo) FindByMessageIDs(_ context.Context, _ []string) (map[string]*entity.MessagePin, error) {
	return r.pins, nil
}

// builderChannelAccess は accessible に含まれるチャンネルだけ参照を許可します
type builderChannelAccess struct {
	service.ChannelAccessService
	accessible map[string]*entity.Channel
}

func (s *builderChannelAccess) AccessibleChannelsByIDs(_ context.Context, channelIDs []string, _ string) (map[string]*entity.Channel, error) {
	result := map[string]*entity.Channel{}
	for _, id := range channelIDs {
		if ch := s.accessible[id]; ch != nil {
			result[id] = ch
		}
	}
	return result, nil
}

const (
	viewerID       = "viewer"
	sourceID       = "source"
	publicTargetID = "public-target"
	secretTargetID = "secret-target"
	deletedID      = "deleted-target"
)

func newTestBuilder() *MessageOutputBuilder {
	deletedAt := time.Now()
	width, height := int32(1280), int32(720)
	return NewMessageOutputBuilder(
		&builderMessageRepo{
			messages: map[string]*entity.Message{
				publicTargetID: {ID: publicTargetID, ChannelID: "public", UserID: "author", Body: strings.Repeat("あ", previewExcerptRunes+10)},
				secretTargetID: {ID: secretTargetID, ChannelID: "secret", UserID: "author", Body: "秘密"},
				deletedID:      {ID: deletedID, ChannelID: "public", UserID: "author", Body: "削除済み", DeletedAt: &deletedAt},
			},
			reactions: map[string][]*entity.MessageReaction{
				sourceID: {{MessageID: sourceID, UserID: "author", Emoji: "👍", CreatedAt: deletedAt}},
			},
		},
		&builderUserRepo{users: []*entity.User{{ID: viewerID, DisplayName: "閲覧者"}, {ID: "author", DisplayName: "投稿者"}}},
		nil,
		&builderUserMentionRepo{},
		&builderGroupMentionRepo{},
		&builderLinkRepo{links: []*entity.MessageLink{
			{ID: "l1", MessageID: sourceID, URL: "https://example.com/public", LinkedMessageID: ptr(publicTargetID)},
			{ID: "l2", MessageID: sourceID, URL: "https://example.com/secret", LinkedMessageID: ptr(secretTargetID)},
			{ID: "l3", MessageID: sourceID, URL: "https://example.com/deleted", LinkedMessageID: ptr(deletedID)},
			{ID: "l4", MessageID: sourceID, URL: "https://example.com/", OGP: entity.OGPData{Title: ptr("外部サイト")}},
		}},
		&builderAttachmentRepo{attachments: map[string][]*entity.Attachment{
			sourceID: {{ID: "a1", MimeType: "image/png", Media: entity.MediaMetadata{Width: &width, Height: &height}}},
		}},
		&builderPinRepo{pins: map[string]*entity.MessagePin{sourceID: {MessageID: sourceID, PinnedBy: viewerID, PinnedAt: deletedAt}}},
		builderPollRepo{},
		&builderChannelAccess{accessible: map[string]*entity.Channel{"public": {ID: "public", Name: "general"}}},
	)
}

func ptr[T any](v T) *T {
	return &v
}

func TestBuildIncludesPreviewsOnlyForAccessibleMessages(t *testing.T) {
	outputs, err := newTestBuilder().Build(context.Background(), viewerID, []*entity.Message{{ID: sourceID, ChannelID: "public", UserID: "author"}})
	if err != nil {
		t.Fatalf("予期しないエラー: %v", err)
	}
	links := outputs[0].Links

	preview := links[0].MessagePreview
	if preview == nil {
		t.Fatal("参照できるメッセージの引用カードがありません")
	}
	if preview.ChannelName != "general" || preview.User.DisplayName != "投稿者" {
		t.Errorf("引用カードのチャンネル名か投稿者が期待と異なります: %+v", preview)
	}
	if got := []rune(preview.BodyExcerpt); len(got) != previewExcerptRunes+1 || got[len(got)-1] != '…' {
		t.Errorf("本文の抜粋が切り詰められていません: %d 文字", len(got))
	}
	if links[1].MessagePreview != nil {
		t.Error("参照できないチャンネルのメッセージが展開されています")
	}
	if links[2].MessagePreview != nil {
		t.Error("削除済みのメッセージが展開されています")
	}
	if links[3].OGP.Title == nil || links[3].MessagePreview != nil {
		t.Error("通常のリンクの OGP が保たれていません")
	}
}

func TestBuildIncludesAppUser(t *testing.T) {
	builder := newTestBuilder()
	builder.userRepo = &builderUserRepo{users: []*entity.User{{ID: "bot", DisplayName: "Deploy", IsApp: true}}}
	outputs, err := builder.Build(context.Background(), viewerID, []*entity.Message{{ID: "m1", ChannelID: "public", UserID: "bot"}})
	if err != nil {
		t.Fatalf("予期しないエラー: %v", err)
	}
	if user := outputs[0].User; user.DisplayName != "Deploy" || !user.IsApp {
		t.Errorf("アプリのユーザーの情報が含まれていません: %+v", user)
	}
}

func TestBuildIncludesPinReactionsAndMedia(t *testing.T) {
	outputs, err := newTestBuilder().Build(context.Background(), viewerID, []*entity.Message{{ID: sourceID, ChannelID: "public", UserID: "author"}})
	if err != nil {
		t.Fatalf("予期しないエラー: %v", err)
	}
	out := outputs[0]

	if out.Pin == nil || out.Pin.PinnedBy.DisplayName != "閲覧者" {
		t.Errorf("ピン留めした人の名前が含まれていません: %+v", out.Pin)
	}
	if len(out.Reactions) != 1 || out.Reactions[0].User.DisplayName != "投稿者" {
		t.Errorf("リアクションしたユーザー名が含まれていません: %+v", out.Reactions)
	}
	if media := out.Attachments[0].Media; media.Width == nil || *media.Width != 1280 {
		t.Errorf("添付ファイルの寸法が含まれていません: %+v", media)
	}
}

type builderPollRepo struct {
	domainrepository.PollRepository
}

func (builderPollRepo) FindByMessageIDs(_ context.Context, ids []string) (map[string]*entity.Poll, error) {
	return map[string]*entity.Poll{sourceID: {ID: "poll", MessageID: sourceID, Options: []entity.PollOption{{ID: "o1"}, {ID: "o2"}}}}, nil
}

func (builderPollRepo) FindVotesByPollIDs(context.Context, []string) ([]*entity.PollVote, error) {
	return []*entity.PollVote{{PollID: "poll", OptionID: "o2", UserID: viewerID}}, nil
}

func TestForBroadcastKeepsOriginal(t *testing.T) {
	outputs, err := newTestBuilder().Build(context.Background(), viewerID, []*entity.Message{{ID: sourceID, ChannelID: "public", UserID: "author"}})
	if err != nil {
		t.Fatalf("予期しないエラー: %v", err)
	}

	stripped := outputs[0].ForBroadcast()

	if stripped.Links[0].MessagePreview != nil {
		t.Error("配信用の出力に引用カードが残っています")
	}
	if stripped.Links[0].LinkedMessageID == nil {
		t.Error("配信用の出力からリンク先のメッセージ ID が消えています")
	}
	if outputs[0].Links[0].MessagePreview == nil {
		t.Error("元の出力の引用カードまで消えています")
	}
	if len(stripped.Poll.MyOptionIDs) != 0 || len(outputs[0].Poll.MyOptionIDs) != 1 || stripped.Poll.Options[1].VoteCount != 1 {
		t.Errorf("配信用の出力に自分の投票が残っているか、集計が消えています: %+v", stripped.Poll)
	}
}

func TestBuildPreview(t *testing.T) {
	builder := newTestBuilder()

	preview, err := builder.BuildPreview(context.Background(), viewerID, publicTargetID)
	if err != nil || preview.MessageID != publicTargetID {
		t.Fatalf("参照できるメッセージの引用カードを取得できません: %v", err)
	}

	for _, id := range []string{secretTargetID, deletedID, "missing"} {
		if _, err := builder.BuildPreview(context.Background(), viewerID, id); !errors.Is(err, domerr.ErrMessageNotFound) {
			t.Errorf("%s は見つからない扱いにすべきです: %v", id, err)
		}
	}
}
