package thread

import (
	"context"
	"reflect"
	"testing"

	"github.com/newt239/chat/internal/domain/entity"
	domainrepository "github.com/newt239/chat/internal/domain/repository"
	"github.com/newt239/chat/internal/usecase/message"
)

type stubListThreadRepo struct {
	domainrepository.ThreadRepository
	output *domainrepository.FindParticipatingThreadsOutput
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

func (stubUserRepo) FindByIDs(_ context.Context, ids []string) ([]*entity.User, error) {
	users := make([]*entity.User, 0, len(ids))
	for _, id := range ids {
		users = append(users, &entity.User{ID: id, DisplayName: "name-" + id})
	}
	return users, nil
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
		stubPinRepo{}, nil,
	)

	out, err := NewThreadLister(threadRepo, builder).ListParticipatingThreads(context.Background(), ListParticipatingThreadsInput{})
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
	if out.Items[0].ReplyCount != 5 {
		t.Errorf("返信数が引き継がれていません: %d", out.Items[0].ReplyCount)
	}
}
