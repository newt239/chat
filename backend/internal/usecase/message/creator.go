package message

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/newt239/chat/internal/domain/entity"
)

// observerTimeout は投稿の処理が終わったあとも外部への通知を続ける上限時間
const observerTimeout = 30 * time.Second

func (i *Interactor) CreateMessage(ctx context.Context, input CreateMessageInput) (*MessageOutput, error) {
	if strings.TrimSpace(input.Body) == "" && len(input.AttachmentIDs) == 0 && input.Location == nil && input.Poll == nil {
		return nil, ErrEmptyMessage
	}
	var poll *entity.Poll
	if input.Poll != nil {
		var err error
		if poll, err = newPoll(input.Poll, time.Now()); err != nil {
			return nil, err
		}
	}
	channel, err := i.channelAccessSvc.EnsureChannelMember(ctx, input.ChannelID, input.UserID)
	if err != nil {
		return nil, err
	}
	parent, err := EnsureReplyTarget(ctx, i.messageRepo, input.ParentID, channel.ID)
	if err != nil {
		return nil, err
	}

	message := &entity.Message{
		ChannelID: channel.ID,
		UserID:    input.UserID,
		ParentID:  input.ParentID,
		Body:      input.Body,
		Location:  input.Location,
	}
	return i.publish(ctx, channel, message, func(txCtx context.Context) error {
		// スレッド返信は親メッセージの投稿者と返信者を自動フォローする
		if parent != nil {
			for _, userID := range []string{parent.UserID, input.UserID} {
				if err := i.threadRepo.FollowThread(txCtx, userID, parent.ID); err != nil {
					return fmt.Errorf("failed to follow thread: %w", err)
				}
			}
		}
		if poll != nil {
			poll.MessageID = message.ID
			if err := i.pollRepo.Create(txCtx, poll); err != nil {
				return fmt.Errorf("failed to create poll: %w", err)
			}
		}
		if len(input.AttachmentIDs) == 0 {
			return nil
		}
		if err := VerifyAttachments(txCtx, i.attachmentRepo, input.UserID, channel.ID, input.AttachmentIDs); err != nil {
			return err
		}
		if err := i.attachmentRepo.AttachToMessage(txCtx, input.AttachmentIDs, message.ID); err != nil {
			return fmt.Errorf("failed to attach files: %w", err)
		}
		return nil
	})
}

// CreateBotMessage はボットユーザー名義のメッセージを投稿します。投稿の可否と返信先は呼び出し側で確認済みであることが前提です
func (i *Interactor) CreateBotMessage(ctx context.Context, channel *entity.Channel, message *entity.Message) (*MessageOutput, error) {
	message.ChannelID = channel.ID
	return i.publish(ctx, channel, message, func(context.Context) error { return nil })
}

// publish はメンションとリンクを解決してからメッセージと一緒に保存し、購読者と外部へ知らせます
func (i *Interactor) publish(ctx context.Context, channel *entity.Channel, message *entity.Message, afterCreate func(txCtx context.Context) error) (*MessageOutput, error) {
	applyBroadcastMentions(message)
	content, err := i.prepareContent(ctx, message.Body, channel.WorkspaceID, nil)
	if err != nil {
		return nil, err
	}

	var result *MessageOutput
	err = i.txManager.Do(ctx, func(txCtx context.Context) error {
		if err := i.messageRepo.Create(txCtx, message); err != nil {
			return fmt.Errorf("failed to create message: %w", err)
		}
		if err := afterCreate(txCtx); err != nil {
			return err
		}
		if err := i.saveContent(txCtx, message.ID, content); err != nil {
			return err
		}
		outputs, err := i.outputBuilder.Build(txCtx, message.UserID, []*entity.Message{message})
		if err != nil {
			return err
		}
		result = &outputs[0]
		return nil
	})
	if err != nil {
		return nil, err
	}

	// 返信が付くと親メッセージの「スレッドあり」も変わる
	indexIDs := []string{message.ID}
	if message.ParentID != nil {
		indexIDs = append(indexIDs, *message.ParentID)
	}
	i.searchIndexer.Sync(ctx, indexIDs...)

	i.notifier.NotifyNewMessage(channel.WorkspaceID, channel.ID, result.ForBroadcast())
	// 投稿の応答を待たせないよう非同期で知らせ、失敗はログに残すだけにする
	for _, observer := range i.observers {
		go func() {
			ctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), observerTimeout)
			defer cancel()
			if err := observer.NotifyNewMessage(ctx, channel, *result); err != nil {
				slog.WarnContext(ctx, "新着メッセージを外部へ知らせられません", "messageId", result.ID, "error", err)
			}
		}()
	}
	return result, nil
}
