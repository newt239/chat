package notification

import (
	"context"
	"fmt"

	"github.com/newt239/chat/internal/domain/entity"
	domainrepository "github.com/newt239/chat/internal/domain/repository"
	"github.com/newt239/chat/internal/domain/service"
	messageuc "github.com/newt239/chat/internal/usecase/message"
)

const maxBodyRunes = 200

// PushMessage は 1 台の端末に送る通知です。Data はウェブの Service Worker とネイティブアプリがそのまま受け取る
type PushMessage struct {
	Token    string
	Platform entity.PushPlatform
	Title    string
	Body     string
	Data     map[string]string
}

// Sender は通知を FCM などへ送り、無効になったトークンを返します
type Sender interface {
	Send(ctx context.Context, messages []PushMessage) (invalidTokens []string, err error)
}

type reason int

const (
	reasonThread reason = iota + 1
	reasonMention
	reasonDM
)

// Dispatcher は新着メッセージの宛先を決めて、その端末へプッシュ通知を送ります
type Dispatcher struct {
	userRepo          domainrepository.UserRepository
	channelMemberRepo domainrepository.ChannelMemberRepository
	channelMuteRepo   domainrepository.ChannelMuteRepository
	threadRepo        domainrepository.ThreadRepository
	pushTokenRepo     domainrepository.PushTokenRepository
	mentionSvc        service.MentionService
	channelAccessSvc  service.ChannelAccessService
	sender            Sender
}

// NewDispatcher の sender が nil のとき（FIREBASE_PROJECT_ID が未設定）は何も送らない
func NewDispatcher(
	userRepo domainrepository.UserRepository,
	channelMemberRepo domainrepository.ChannelMemberRepository,
	channelMuteRepo domainrepository.ChannelMuteRepository,
	threadRepo domainrepository.ThreadRepository,
	pushTokenRepo domainrepository.PushTokenRepository,
	mentionSvc service.MentionService,
	channelAccessSvc service.ChannelAccessService,
	sender Sender,
) *Dispatcher {
	return &Dispatcher{
		userRepo:          userRepo,
		channelMemberRepo: channelMemberRepo,
		channelMuteRepo:   channelMuteRepo,
		threadRepo:        threadRepo,
		pushTokenRepo:     pushTokenRepo,
		mentionSvc:        mentionSvc,
		channelAccessSvc:  channelAccessSvc,
		sender:            sender,
	}
}

func (d *Dispatcher) NotifyNewMessage(ctx context.Context, channel *entity.Channel, message messageuc.MessageOutput) error {
	if d.sender == nil {
		return nil
	}
	candidates, err := d.candidates(ctx, channel, message)
	if err != nil {
		return err
	}
	recipients, err := d.filterRecipients(ctx, channel, candidates)
	if err != nil || len(recipients) == 0 {
		return err
	}

	tokens, err := d.pushTokenRepo.FindByUserIDs(ctx, recipients)
	if err != nil {
		return fmt.Errorf("failed to load push tokens: %w", err)
	}
	if len(tokens) == 0 {
		return nil
	}

	body, err := d.mentionSvc.RenderPlain(ctx, message.Body)
	if err != nil {
		return fmt.Errorf("failed to render message body: %w", err)
	}
	title, body, data := content(channel, message, body)
	messages := make([]PushMessage, 0, len(tokens))
	for _, t := range tokens {
		messages = append(messages, PushMessage{Token: t.Token, Platform: t.Platform, Title: title, Body: body, Data: data})
	}
	invalid, err := d.sender.Send(ctx, messages)
	if len(invalid) > 0 {
		if delErr := d.pushTokenRepo.DeleteTokens(ctx, invalid); delErr != nil {
			return fmt.Errorf("failed to delete invalid push tokens: %w", delErr)
		}
	}
	return err
}

// candidates は通知の理由ごとに宛先の候補を集めます。複数に当てはまるときは強い理由を残す
func (d *Dispatcher) candidates(ctx context.Context, channel *entity.Channel, message messageuc.MessageOutput) (map[string]reason, error) {
	result := map[string]reason{}
	add := func(userID string, r reason) {
		if r > result[userID] {
			result[userID] = r
		}
	}

	if message.ParentID != nil {
		followers, err := d.threadRepo.FindFollowerIDs(ctx, *message.ParentID)
		if err != nil {
			return nil, fmt.Errorf("failed to load thread followers: %w", err)
		}
		for _, id := range followers {
			add(id, reasonThread)
		}
	}
	// グループへのメンションは投稿時点のメンバーに展開済み
	for _, m := range message.Mentions {
		add(m.UserID, reasonMention)
	}
	if channel.IsDM() {
		members, err := d.channelMemberRepo.FindMembers(ctx, channel.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to load DM members: %w", err)
		}
		for _, m := range members {
			add(m.UserID, reasonDM)
		}
	}

	delete(result, message.UserID)
	return result, nil
}

// filterRecipients は通知設定・閲覧権限・ミュートで宛先を絞ります
func (d *Dispatcher) filterRecipients(ctx context.Context, channel *entity.Channel, candidates map[string]reason) ([]string, error) {
	if len(candidates) == 0 {
		return nil, nil
	}
	ids := make([]string, 0, len(candidates))
	for id := range candidates {
		ids = append(ids, id)
	}
	users, err := d.userRepo.FindByIDs(ctx, ids)
	if err != nil {
		return nil, fmt.Errorf("failed to load recipients: %w", err)
	}
	accessible, err := d.channelAccessSvc.FilterUsersWithAccess(ctx, channel, ids)
	if err != nil {
		return nil, err
	}
	muted, err := d.channelMuteRepo.FindMutedUserIDs(ctx, channel.ID, ids)
	if err != nil {
		return nil, fmt.Errorf("failed to load mutes: %w", err)
	}

	recipients := []string{}
	for _, u := range users {
		if !u.IsApp && wants(u.Preferences.NotificationLevel, candidates[u.ID]) && accessible[u.ID] && !muted[u.ID] {
			recipients = append(recipients, u.ID)
		}
	}
	return recipients, nil
}

// wants は「すべて」ならフォロー中のスレッドへの返信も、「メンションと DM のみ」ならそれだけを通知します
func wants(level entity.NotificationLevel, r reason) bool {
	switch level {
	case entity.NotificationLevelAll:
		return true
	case entity.NotificationLevelMentions:
		return r >= reasonMention
	default:
		return false
	}
}

// content の body は ID 記法を名前に置き換えた本文です
func content(channel *entity.Channel, message messageuc.MessageOutput, body string) (string, string, map[string]string) {
	title := message.User.DisplayName
	if !channel.IsDM() {
		title += " · #" + channel.Name
	}
	if runes := []rune(body); len(runes) > maxBodyRunes {
		body = string(runes[:maxBodyRunes]) + "…"
	}
	data := map[string]string{
		"workspaceId": channel.WorkspaceID,
		"channelId":   channel.ID,
		"messageId":   message.ID,
		"link":        entity.MessagePermalinkPath(channel.WorkspaceID, channel.ID, message.ID, message.ParentID),
	}
	return title, body, data
}
