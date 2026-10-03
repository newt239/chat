package poll

import (
	"context"
	"errors"
	"slices"
	"testing"
	"time"

	"github.com/newt239/chat/internal/domain/entity"
	domerr "github.com/newt239/chat/internal/domain/errors"
	domainrepository "github.com/newt239/chat/internal/domain/repository"
	domainservice "github.com/newt239/chat/internal/domain/service"
	messageuc "github.com/newt239/chat/internal/usecase/message"
)

const (
	pollID    = "poll"
	messageID = "message"
	authorID  = "author"
	voterID   = "voter"
)

type fakePollRepo struct {
	domainrepository.PollRepository
	poll *entity.Poll
}

func (r *fakePollRepo) FindByID(context.Context, string) (*entity.Poll, error) {
	copied := *r.poll
	return &copied, nil
}

func (r *fakePollRepo) Close(_ context.Context, _ string, at time.Time) error {
	r.poll.ClosedAt = &at
	return nil
}

type stubMessageRepo struct {
	domainrepository.MessageRepository
}

func (stubMessageRepo) FindByID(context.Context, string) (*entity.Message, error) {
	return &entity.Message{ID: messageID, ChannelID: "channel", UserID: authorID}, nil
}

type stubWorkspaceRepo struct {
	domainrepository.WorkspaceRepository
}

func (stubWorkspaceRepo) FindMember(_ context.Context, _ string, userID string) (*entity.WorkspaceMember, error) {
	return &entity.WorkspaceMember{UserID: userID, Role: entity.WorkspaceRoleMember}, nil
}

type stubAccess struct {
	domainservice.ChannelAccessService
}

func (stubAccess) EnsureChannelMember(_ context.Context, id string, _ string) (*entity.Channel, error) {
	return &entity.Channel{ID: id, WorkspaceID: "ws"}, nil
}

type stubTx struct{}

func (stubTx) Do(ctx context.Context, fn func(context.Context) error) error { return fn(ctx) }

type stubNotifier struct {
	messageuc.Notifier
	updated int
}

func (n *stubNotifier) NotifyUpdatedMessage(string, string, messageuc.MessageOutput) {
	n.updated++
}

// 出力の組み立て役は渡さず、組み立ての前に失敗する場合だけを確かめる
func newInteractor(poll *entity.Poll) (*Interactor, *fakePollRepo) {
	repo := &fakePollRepo{poll: poll}
	return NewInteractor(repo, stubMessageRepo{}, stubWorkspaceRepo{}, stubAccess{}, nil, &stubNotifier{}, stubTx{}), repo
}

func TestVoteValidates(t *testing.T) {
	past := time.Now().Add(-time.Minute)
	single := &entity.Poll{ID: pollID, MessageID: messageID, Options: []entity.PollOption{{ID: "a"}, {ID: "b"}}}
	tests := []struct {
		name      string
		poll      *entity.Poll
		optionIDs []string
		want      error
	}{
		{name: "単一選択で 2 つは選べない", poll: single, optionIDs: []string{"a", "b"}, want: ErrInvalidVote},
		{name: "ほかの投票の選択肢は選べない", poll: single, optionIDs: []string{"x"}, want: ErrInvalidVote},
		{name: "締切を過ぎたら投票できない", poll: &entity.Poll{ID: pollID, MessageID: messageID, ClosesAt: &past, Options: single.Options}, optionIDs: []string{"a"}, want: ErrPollClosed},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			i, _ := newInteractor(tt.poll)
			if _, err := i.Vote(context.Background(), VoteInput{PollID: pollID, UserID: voterID, OptionIDs: tt.optionIDs}); !errors.Is(err, tt.want) {
				t.Fatalf("got=%v want=%v", err, tt.want)
			}
		})
	}
}

func TestNormalizeChoiceRemovesDuplicates(t *testing.T) {
	poll := &entity.Poll{AllowMultiple: true, Options: []entity.PollOption{{ID: "a"}, {ID: "b"}}}
	got, err := normalizeChoice(poll, []string{"b", "a", "b"})
	if err != nil || !slices.Equal(got, []string{"a", "b"}) {
		t.Fatalf("got=%v err=%v", got, err)
	}
	if got, err := normalizeChoice(poll, nil); err != nil || len(got) != 0 {
		t.Fatalf("空の選択で取り消せません: %v %v", got, err)
	}
}

func TestCloseRequiresAuthorOrAdmin(t *testing.T) {
	i, _ := newInteractor(&entity.Poll{ID: pollID, MessageID: messageID})
	if _, err := i.Close(context.Background(), CloseInput{PollID: pollID, UserID: voterID}); !errors.Is(err, domerr.ErrUnauthorized) {
		t.Fatalf("作成者と管理者以外の締め切りを拒否していません: %v", err)
	}
}
