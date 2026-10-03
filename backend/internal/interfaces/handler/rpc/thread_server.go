package rpc

import (
	"context"

	domainrepository "github.com/newt239/chat/internal/domain/repository"
	chatv1 "github.com/newt239/chat/internal/gen/chat/v1"
	"github.com/newt239/chat/internal/interfaces/presenter"
	messageuc "github.com/newt239/chat/internal/usecase/message"
	threaduc "github.com/newt239/chat/internal/usecase/thread"
)

type ThreadServer struct {
	MessageUC *messageuc.Interactor
	UC        *threaduc.Interactor
}

func (s *ThreadServer) GetThreadReplies(ctx context.Context, req *chatv1.GetThreadRepliesRequest) (*chatv1.GetThreadRepliesResponse, error) {
	out, err := s.MessageUC.GetThreadReplies(ctx, messageuc.GetThreadRepliesInput{
		MessageID:     req.MessageId,
		UserID:        userIDFrom(ctx),
		Limit:         int(req.Limit),
		Since:         optionalTime(req.GetSince()),
		Until:         optionalTime(req.GetUntil()),
		AroundReplyID: req.AroundReplyId,
	})
	if err != nil {
		return nil, err
	}
	return &chatv1.GetThreadRepliesResponse{
		ParentMessage: presenter.Message(out.ParentMessage),
		Replies:       presenter.ConvertAll(out.Replies, presenter.Message),
		HasMore:       out.HasMore,
		HasNewer:      out.HasNewer,
		ReplyCount:    int32(out.ReplyCount),
	}, nil
}

func (s *ThreadServer) GetThreadMetadata(ctx context.Context, req *chatv1.GetThreadMetadataRequest) (*chatv1.GetThreadMetadataResponse, error) {
	out, err := s.MessageUC.GetThreadMetadata(ctx, messageuc.MessageInput{MessageID: req.MessageId, UserID: userIDFrom(ctx)})
	if err != nil {
		return nil, err
	}
	return &chatv1.GetThreadMetadataResponse{Metadata: presenter.ThreadMetadata(*out)}, nil
}

func (s *ThreadServer) ListParticipatingThreads(ctx context.Context, req *chatv1.ListParticipatingThreadsRequest) (*chatv1.ListParticipatingThreadsResponse, error) {
	input := domainrepository.FindParticipatingThreadsInput{WorkspaceID: req.WorkspaceId, UserID: userIDFrom(ctx), Limit: int(req.Limit)}
	if c := req.Cursor; c != nil {
		input.Cursor = &domainrepository.ThreadCursor{LastActivityAt: c.LastActivityAt.AsTime(), ThreadID: c.ThreadId}
	}
	out, err := s.UC.ListParticipatingThreads(ctx, input)
	if err != nil {
		return nil, err
	}
	return presenter.ParticipatingThreads(out), nil
}

func (s *ThreadServer) MarkThreadRead(ctx context.Context, req *chatv1.MarkThreadReadRequest) (*chatv1.MarkThreadReadResponse, error) {
	return &chatv1.MarkThreadReadResponse{}, s.UC.MarkThreadRead(ctx, req.ThreadId, userIDFrom(ctx))
}

func (s *ThreadServer) FollowThread(ctx context.Context, req *chatv1.FollowThreadRequest) (*chatv1.FollowThreadResponse, error) {
	return &chatv1.FollowThreadResponse{}, s.UC.SetFollowing(ctx, req.MessageId, userIDFrom(ctx), true)
}

func (s *ThreadServer) UnfollowThread(ctx context.Context, req *chatv1.UnfollowThreadRequest) (*chatv1.UnfollowThreadResponse, error) {
	return &chatv1.UnfollowThreadResponse{}, s.UC.SetFollowing(ctx, req.MessageId, userIDFrom(ctx), false)
}
