// Package poll はメッセージに付けた投票への投票と締め切りを扱います
package poll

import (
	"context"
	"fmt"
	"slices"
	"time"

	"github.com/newt239/chat/internal/domain/entity"
	domerr "github.com/newt239/chat/internal/domain/errors"
	domainrepository "github.com/newt239/chat/internal/domain/repository"
	domainservice "github.com/newt239/chat/internal/domain/service"
	domaintransaction "github.com/newt239/chat/internal/domain/transaction"
	messageuc "github.com/newt239/chat/internal/usecase/message"
)

var (
	ErrPollNotFound = domerr.New(domerr.ErrNotFound, "投票が見つかりません")
	ErrPollClosed   = domerr.New(domerr.ErrFailedPrecondition, "この投票は締め切られています")
	ErrInvalidVote  = domerr.New(domerr.ErrValidation, "選択肢が正しくありません")
)

// PollInput は締め切りでは OptionIDs を使いません。投票で空にすると投票を取り消します
type PollInput struct {
	PollID    string
	UserID    string
	OptionIDs []string
}

type Interactor struct {
	pollRepo         domainrepository.PollRepository
	messageRepo      domainrepository.MessageRepository
	workspaceRepo    domainrepository.WorkspaceRepository
	channelAccessSvc domainservice.ChannelAccessService
	outputBuilder    *messageuc.MessageOutputBuilder
	notifier         messageuc.Notifier
	txManager        domaintransaction.Manager
}

func New(
	pollRepo domainrepository.PollRepository,
	messageRepo domainrepository.MessageRepository,
	workspaceRepo domainrepository.WorkspaceRepository,
	channelAccessSvc domainservice.ChannelAccessService,
	outputBuilder *messageuc.MessageOutputBuilder,
	notifier messageuc.Notifier,
	txManager domaintransaction.Manager,
) *Interactor {
	return &Interactor{
		pollRepo:         pollRepo,
		messageRepo:      messageRepo,
		workspaceRepo:    workspaceRepo,
		channelAccessSvc: channelAccessSvc,
		outputBuilder:    outputBuilder,
		notifier:         notifier,
		txManager:        txManager,
	}
}

// Vote は自分の票を選んだ選択肢に置き換え、集計を更新したメッセージを配信します
func (i *Interactor) Vote(ctx context.Context, input PollInput) (*messageuc.MessageOutput, error) {
	poll, message, ch, err := i.find(ctx, input.PollID, input.UserID)
	if err != nil {
		return nil, err
	}
	if poll.IsClosed(time.Now()) {
		return nil, ErrPollClosed
	}
	optionIDs, err := normalizeChoice(poll, input.OptionIDs)
	if err != nil {
		return nil, err
	}
	if err := i.txManager.Do(ctx, func(txCtx context.Context) error {
		return i.pollRepo.ReplaceVotes(txCtx, poll.ID, input.UserID, optionIDs)
	}); err != nil {
		return nil, fmt.Errorf("failed to save votes: %w", err)
	}
	return i.publish(ctx, ch, message, input.UserID)
}

// normalizeChoice は重複を除き、選択肢がこの投票のもので、単一選択なら 1 つ以下であることを確かめます
func normalizeChoice(poll *entity.Poll, optionIDs []string) ([]string, error) {
	unique := slices.Compact(slices.Sorted(slices.Values(optionIDs)))
	if !poll.AllowMultiple && len(unique) > 1 {
		return nil, ErrInvalidVote
	}
	for _, id := range unique {
		if !poll.HasOption(id) {
			return nil, ErrInvalidVote
		}
	}
	return unique, nil
}

// Close は投票を締め切ります。作成者とワークスペースの管理者だけができる
func (i *Interactor) Close(ctx context.Context, input PollInput) (*messageuc.MessageOutput, error) {
	poll, message, ch, err := i.find(ctx, input.PollID, input.UserID)
	if err != nil {
		return nil, err
	}
	if message.UserID != input.UserID {
		if _, err := domainservice.EnsureAdmin(ctx, i.workspaceRepo, ch.WorkspaceID, input.UserID); err != nil {
			return nil, err
		}
	}
	if poll.IsClosed(time.Now()) {
		return nil, ErrPollClosed
	}
	if err := i.pollRepo.Close(ctx, poll.ID, time.Now()); err != nil {
		return nil, fmt.Errorf("failed to close poll: %w", err)
	}
	return i.publish(ctx, ch, message, input.UserID)
}

// find は投票と、それを付けたメッセージ・チャンネルを返します。チャンネルに参加している人だけが操作できる
func (i *Interactor) find(ctx context.Context, pollID, userID string) (*entity.Poll, *entity.Message, *entity.Channel, error) {
	poll, err := i.pollRepo.FindByID(ctx, pollID)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to load poll: %w", err)
	}
	if poll == nil {
		return nil, nil, nil, ErrPollNotFound
	}
	message, err := i.messageRepo.FindByID(ctx, poll.MessageID)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to load message: %w", err)
	}
	if message == nil || message.DeletedAt != nil {
		return nil, nil, nil, ErrPollNotFound
	}
	ch, err := i.channelAccessSvc.EnsureChannelMember(ctx, message.ChannelID, userID)
	if err != nil {
		return nil, nil, nil, err
	}
	return poll, message, ch, nil
}

// publish は操作した人から見たメッセージを返し、自分の投票を除いた集計をチャンネルに配信します
func (i *Interactor) publish(ctx context.Context, ch *entity.Channel, message *entity.Message, userID string) (*messageuc.MessageOutput, error) {
	outputs, err := i.outputBuilder.Build(ctx, userID, []*entity.Message{message})
	if err != nil {
		return nil, err
	}
	i.notifier.NotifyUpdatedMessage(ch.WorkspaceID, ch.ID, outputs[0].ForBroadcast())
	return &outputs[0], nil
}
