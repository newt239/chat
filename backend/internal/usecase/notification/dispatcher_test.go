package notification

import (
	"context"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/newt239/chat/internal/domain/entity"
	domainrepository "github.com/newt239/chat/internal/domain/repository"
	"github.com/newt239/chat/internal/domain/service"
	messageuc "github.com/newt239/chat/internal/usecase/message"
)

type stubUserRepo struct {
	domainrepository.UserRepository
	levels map[string]entity.NotificationLevel
}

func (r stubUserRepo) FindByIDs(_ context.Context, ids []string) (map[string]*entity.User, error) {
	users := map[string]*entity.User{}
	for _, id := range ids {
		users[id] = &entity.User{ID: id, Preferences: entity.UserPreferences{NotificationLevel: r.levels[id]}}
	}
	return users, nil
}

type stubMemberRepo struct {
	domainrepository.ChannelMemberRepository
	members []string
}

func (r stubMemberRepo) FindMembersByChannelIDs(context.Context, []string) ([]*entity.ChannelMember, error) {
	result := []*entity.ChannelMember{}
	for _, id := range r.members {
		result = append(result, &entity.ChannelMember{UserID: id})
	}
	return result, nil
}

type stubMuteRepo struct {
	domainrepository.ChannelMuteRepository
	mutedBy string
}

func (r stubMuteRepo) FindMutedUserIDs(context.Context, string, []string) (map[string]bool, error) {
	return map[string]bool{r.mutedBy: true}, nil
}

type stubThreadRepo struct {
	domainrepository.ThreadRepository
	followers []string
}

func (r stubThreadRepo) FindFollowerIDs(context.Context, string) ([]string, error) {
	return r.followers, nil
}

type stubMentionService struct {
	service.MentionService
}

func (stubMentionService) RenderPlain(_ context.Context, body string) (string, error) {
	return strings.ReplaceAll(body, "<@u1>", "@bob"), nil
}

type stubTokenRepo struct {
	domainrepository.PushTokenRepository
	deleted []string
}

func (r *stubTokenRepo) FindByUserIDs(_ context.Context, userIDs []string) ([]*entity.PushToken, error) {
	result := []*entity.PushToken{}
	for _, id := range userIDs {
		result = append(result, &entity.PushToken{UserID: id, Token: "token-" + id, Platform: entity.PushPlatformWeb})
	}
	return result, nil
}

func (r *stubTokenRepo) DeleteTokens(_ context.Context, tokens []string) error {
	r.deleted = append(r.deleted, tokens...)
	return nil
}

type stubAccess struct {
	service.ChannelAccessService
	denied string
}

func (s stubAccess) FilterUsersWithAccess(_ context.Context, _ *entity.Channel, userIDs []string) (map[string]bool, error) {
	result := map[string]bool{}
	for _, id := range userIDs {
		if id != s.denied {
			result[id] = true
		}
	}
	return result, nil
}

type stubSender struct {
	sent    []PushMessage
	invalid []string
}

func (s *stubSender) Send(_ context.Context, messages []PushMessage) ([]string, error) {
	s.sent = append(s.sent, messages...)
	return s.invalid, nil
}

type fixture struct {
	levels    map[string]entity.NotificationLevel
	members   []string
	followers []string
	mutedBy   string
	denied    string
}

func (f fixture) run(t *testing.T, channel *entity.Channel, message messageuc.MessageOutput) (*stubSender, *stubTokenRepo) {
	t.Helper()
	sender := &stubSender{invalid: []string{"token-stale"}}
	tokens := &stubTokenRepo{}
	d := NewDispatcher(
		stubUserRepo{levels: f.levels},
		stubMemberRepo{members: f.members},
		stubMuteRepo{mutedBy: f.mutedBy},
		stubThreadRepo{followers: f.followers},
		tokens,
		stubMentionService{},
		stubAccess{denied: f.denied},
		sender,
	)
	if err := d.NotifyNewMessage(context.Background(), channel, message); err != nil {
		t.Fatalf("送信に失敗しました: %v", err)
	}
	return sender, tokens
}

func recipients(sent []PushMessage) []string {
	result := []string{}
	for _, m := range sent {
		result = append(result, m.Token)
	}
	slices.Sort(result)
	return result
}

func TestDispatchChannelMessage(t *testing.T) {
	parentID := "parent"
	channel := &entity.Channel{ID: "c1", WorkspaceID: "ws", Name: "general", Type: entity.ChannelTypePublic}
	message := messageuc.MessageOutput{
		ID:       "m1",
		UserID:   "sender",
		User:     messageuc.UserInfo{DisplayName: "Alice"},
		ParentID: &parentID,
		Body:     "hello <@u1>",
		// グループへのメンションは投稿時点のメンバーに展開済み
		Mentions: []messageuc.UserMention{{UserID: "mentioned"}, {UserID: "muted"}, {UserID: "outsider"}, {UserID: "grouped"}, {UserID: "silent"}},
	}
	f := fixture{
		levels: map[string]entity.NotificationLevel{
			"sender": entity.NotificationLevelAll, "mentioned": entity.NotificationLevelMentions,
			"muted": entity.NotificationLevelAll, "outsider": entity.NotificationLevelAll,
			"follower-all": entity.NotificationLevelAll, "follower-mentions": entity.NotificationLevelMentions,
			"grouped": entity.NotificationLevelMentions, "silent": entity.NotificationLevelNone,
		},
		followers: []string{"sender", "follower-all", "follower-mentions"},
		mutedBy:   "muted",
		denied:    "outsider",
	}

	sender, tokens := f.run(t, channel, message)

	want := []string{"token-follower-all", "token-grouped", "token-mentioned"}
	if got := recipients(sender.sent); !reflect.DeepEqual(got, want) {
		t.Errorf("宛先が期待と異なります: got=%v want=%v", got, want)
	}
	first := sender.sent[0]
	if first.Title != "Alice · #general" || first.Body != "hello @bob" || first.Data["link"] != "/app/ws/c1/thread/parent?message=m1" {
		t.Errorf("通知の内容が期待と異なります: %+v", first)
	}
	if !reflect.DeepEqual(tokens.deleted, []string{"token-stale"}) {
		t.Errorf("無効なトークンが削除されていません: %v", tokens.deleted)
	}
}

func TestDispatchDM(t *testing.T) {
	channel := &entity.Channel{ID: "d1", WorkspaceID: "ws", Type: entity.ChannelTypeDM}
	message := messageuc.MessageOutput{ID: "m1", UserID: "sender", User: messageuc.UserInfo{DisplayName: "Alice"}, Body: "hi"}
	f := fixture{
		levels:  map[string]entity.NotificationLevel{"sender": entity.NotificationLevelAll, "bob": entity.NotificationLevelMentions},
		members: []string{"sender", "bob"},
	}

	sender, _ := f.run(t, channel, message)

	if got := recipients(sender.sent); !reflect.DeepEqual(got, []string{"token-bob"}) {
		t.Errorf("DM の相手だけに送られていません: %v", got)
	}
	if sender.sent[0].Title != "Alice" {
		t.Errorf("DM の題名は送信者名だけにする: %q", sender.sent[0].Title)
	}
}

func TestNotifyNewMessageWithoutSenderDoesNothing(t *testing.T) {
	d := NewDispatcher(nil, nil, nil, nil, nil, nil, nil, nil)
	if err := d.NotifyNewMessage(context.Background(), &entity.Channel{}, messageuc.MessageOutput{}); err != nil {
		t.Fatal(err)
	}
}
