package thread

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/newt239/chat/internal/domain/entity"
	domerr "github.com/newt239/chat/internal/domain/errors"
	domainrepository "github.com/newt239/chat/internal/domain/repository"
	"github.com/newt239/chat/internal/domain/service"
	"github.com/newt239/chat/internal/usecase/message"
)

type stubListThreadRepo struct {
	domainrepository.ThreadRepository
	output *domainrepository.FindParticipatingThreadsOutput
}

func (r *stubListThreadRepo) FindFollowedThreadIDs(context.Context, string, []string) (map[string]bool, error) {
	return map[string]bool{"t2": true}, nil
}

func (r *stubListThreadRepo) FindParticipatingThreads(_ context.Context, _ domainrepository.FindParticipatingThreadsInput) (*domainrepository.FindParticipatingThreadsOutput, error) {
	return r.output, nil
}

type stubReactionRepo struct {
	domainrepository.MessageRepository
}

func (stubReactionRepo) FindReactionsByMessageIDs(_ context.Context, _ []string) (map[string][]*entity.MessageReaction, error) {
	return map[string][]*entity.MessageReaction{}, nil
}

type stubUserRepo struct {
	domainrepository.UserRepository
}

func (stubUserRepo) FindByIDs(_ context.Context, ids []string) (map[string]*entity.User, error) {
	users := make(map[string]*entity.User, len(ids))
	for _, id := range ids {
		users[id] = &entity.User{ID: id, DisplayName: "name-" + id}
	}
	return users, nil
}

type stubWorkspaceRepo struct {
	domainrepository.WorkspaceRepository
}

func (stubWorkspaceRepo) FindMember(_ context.Context, workspaceID, userID string) (*entity.WorkspaceMember, error) {
	return &entity.WorkspaceMember{WorkspaceID: workspaceID, UserID: userID}, nil
}

type stubUserMentionRepo struct {
	domainrepository.MessageUserMentionRepository
}

func (stubUserMentionRepo) FindByMessageIDs(_ context.Context, _ []string) ([]*entity.MessageUserMention, error) {
	return nil, nil
}

type stubGroupMentionRepo struct {
	domainrepository.MessageGroupMentionRepository
}

func (stubGroupMentionRepo) FindByMessageIDs(_ context.Context, _ []string) ([]*entity.MessageGroupMention, error) {
	return nil, nil
}

type stubLinkRepo struct {
	domainrepository.MessageLinkRepository
}

func (stubLinkRepo) FindByMessageIDs(_ context.Context, _ []string) ([]*entity.MessageLink, error) {
	return nil, nil
}

type stubPinRepo struct {
	domainrepository.PinRepository
}

func (stubPinRepo) FindByMessageIDs(_ context.Context, _ []string) (map[string]*entity.MessagePin, error) {
	return map[string]*entity.MessagePin{}, nil
}

type stubAttachmentRepo struct {
	domainrepository.AttachmentRepository
}

func (stubAttachmentRepo) FindByMessageIDs(_ context.Context, _ []string) (map[string][]*entity.Attachment, error) {
	return map[string][]*entity.Attachment{}, nil
}

func TestListParticipatingThreadsBuildsMessages(t *testing.T) {
	msg := func(id, userID string) *entity.Message { return &entity.Message{ID: id, UserID: userID} }
	threadRepo := &stubListThreadRepo{output: &domainrepository.FindParticipatingThreadsOutput{Items: []domainrepository.ParticipatingThread{
		{ThreadID: "t1", FirstMessage: msg("t1", "u1"), LatestReplies: []*entity.Message{msg("r1", "u2"), msg("r2", "u3")}, ReplyCount: 5},
		{ThreadID: "t2", FirstMessage: msg("t2", "u2"), LatestReplies: []*entity.Message{}},
		{ThreadID: "t3", FirstMessage: msg("t3", "u3"), LatestReplies: []*entity.Message{msg("r3", "u1")}},
	}}}
	builder := message.NewMessageOutputBuilder(
		stubReactionRepo{}, stubUserRepo{}, nil, stubUserMentionRepo{}, stubGroupMentionRepo{}, stubLinkRepo{}, stubAttachmentRepo{},
		stubPinRepo{}, stubPollRepo{}, nil,
	)

	out, err := New(threadRepo, stubWorkspaceRepo{}, nil, builder).ListParticipatingThreads(context.Background(), domainrepository.FindParticipatingThreadsInput{})
	if err != nil {
		t.Fatalf("取得に失敗しました: %v", err)
	}

	want := map[string][]string{"t1": {"r1", "r2"}, "t2": {}, "t3": {"r3"}}
	for _, item := range out.Items {
		if item.FirstMessage.ID != item.ThreadID || item.FirstMessage.User.DisplayName == "" {
			t.Errorf("親メッセージが正しく組み立てられていません: %+v", item.FirstMessage)
		}
		got := []string{}
		for _, r := range item.LatestReplies {
			got = append(got, r.ID)
		}
		if !reflect.DeepEqual(got, want[item.ThreadID]) {
			t.Errorf("%s の最新の返信が期待と異なります: got=%v want=%v", item.ThreadID, got, want[item.ThreadID])
		}
	}
	if out.Items[0].IsFollowing || !out.Items[1].IsFollowing {
		t.Errorf("閲覧者のフォロー状態が入っていません: %v %v", out.Items[0].IsFollowing, out.Items[1].IsFollowing)
	}
	if out.Items[0].ReplyCount != 5 {
		t.Errorf("返信数が引き継がれていません: %d", out.Items[0].ReplyCount)
	}
}

type stubPollRepo struct {
	domainrepository.PollRepository
}

func (stubPollRepo) FindByMessageIDs(context.Context, []string) (map[string]*entity.Poll, error) {
	return map[string]*entity.Poll{}, nil
}

func (stubPollRepo) FindVotesByPollIDs(context.Context, []string) ([]*entity.PollVote, error) {
	return []*entity.PollVote{}, nil
}

type stubChannelAccessService struct {
	service.ChannelAccessService
	err error
}

func (s *stubChannelAccessService) EnsureMessageAccess(_ context.Context, messageID, _ string) (*entity.Message, *entity.Channel, error) {
	if s.err != nil {
		return nil, nil, s.err
	}
	return &entity.Message{ID: messageID, ChannelID: "ch1"}, &entity.Channel{ID: "ch1"}, nil
}

type stubThreadRepo struct {
	domainrepository.ThreadRepository
	upsertCalls int
	followCalls int
}

func (r *stubThreadRepo) UpsertReadState(context.Context, string, string, time.Time) error {
	r.upsertCalls++
	return nil
}

func (r *stubThreadRepo) FollowThread(context.Context, string, string) error {
	r.followCalls++
	return nil
}

func TestThreadOperationsRequireMessageAccess(t *testing.T) {
	for _, accessErr := range []error{domerr.ErrMessageNotFound, domerr.ErrUnauthorized} {
		threadRepo := &stubThreadRepo{}
		uc := New(threadRepo, nil, &stubChannelAccessService{err: accessErr}, nil)

		if err := uc.MarkThreadRead(context.Background(), "t1", "u1"); !errors.Is(err, accessErr) {
			t.Errorf("見られないスレッドの既読が拒否されていません: %v", err)
		}
		if err := uc.SetFollowing(context.Background(), "t1", "u1", true); !errors.Is(err, accessErr) {
			t.Errorf("見られないスレッドのフォローが拒否されていません: %v", err)
		}
		if threadRepo.upsertCalls != 0 || threadRepo.followCalls != 0 {
			t.Errorf("拒否したのに更新されています: %+v", threadRepo)
		}
	}
}

func TestMarkThreadReadSucceeds(t *testing.T) {
	threadRepo := &stubThreadRepo{}
	uc := New(threadRepo, nil, &stubChannelAccessService{}, nil)

	if err := uc.MarkThreadRead(context.Background(), "t1", "u1"); err != nil {
		t.Fatalf("既読更新に失敗しました: %v", err)
	}
	if threadRepo.upsertCalls != 1 {
		t.Errorf("既読が更新されていません: %d 回", threadRepo.upsertCalls)
	}
}
