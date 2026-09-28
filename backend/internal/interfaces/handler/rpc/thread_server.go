package rpc

import (
	"context"

	chatv1 "github.com/newt239/chat/internal/gen/chat/v1"
	"github.com/newt239/chat/internal/interfaces/presenter"
	messageuc "github.com/newt239/chat/internal/usecase/message"
	threaduc "github.com/newt239/chat/internal/usecase/thread"
)

const defaultThreadLimit = 20

type ThreadServer struct {
	MessageUC    messageuc.MessageUseCase
	ThreadLister *threaduc.ThreadLister
	ThreadReader *threaduc.ThreadReader
}

func (s *ThreadServer) GetThreadReplies(ctx context.Context, req *chatv1.GetThreadRepliesRequest) (*chatv1.GetThreadRepliesResponse, error) {
	out, err := s.MessageUC.GetThreadReplies(ctx, messageuc.GetThreadRepliesInput{MessageID: req.MessageId, UserID: userIDFrom(ctx)})
	if err != nil {
		return nil, err
	}
	return &chatv1.GetThreadRepliesResponse{
		ParentMessage: presenter.Message(out.ParentMessage),
		Replies:       presenter.ConvertAll(out.Replies, presenter.Message),
		HasMore:       out.HasMore,
	}, nil
}

func (s *ThreadServer) GetThreadMetadata(ctx context.Context, req *chatv1.GetThreadMetadataRequest) (*chatv1.GetThreadMetadataResponse, error) {
	out, err := s.MessageUC.GetThreadMetadata(ctx, messageuc.GetThreadMetadataInput{MessageID: req.MessageId, UserID: userIDFrom(ctx)})
	if err != nil {
		return nil, err
	}
	return &chatv1.GetThreadMetadataResponse{Metadata: presenter.ThreadMetadata(*out)}, nil
}

func (s *ThreadServer) ListParticipatingThreads(ctx context.Context, req *chatv1.ListParticipatingThreadsRequest) (*chatv1.ListParticipatingThreadsResponse, error) {
	input := threaduc.ListParticipatingThreadsInput{WorkspaceID: req.WorkspaceId, UserID: userIDFrom(ctx), Limit: defaultThreadLimit}
	if req.Limit > 0 {
		input.Limit = int(req.Limit)
	}
	if req.Cursor != nil {
		input.CursorLastActivityAt = optionalTime(req.Cursor.LastActivityAt)
		input.CursorThreadID = &req.Cursor.ThreadId
	}
	out, err := s.ThreadLister.ListParticipatingThreads(ctx, input)
	if err != nil {
		return nil, err
	}
	return presenter.ParticipatingThreads(out), nil
}

func (s *ThreadServer) MarkThreadRead(ctx context.Context, req *chatv1.MarkThreadReadRequest) (*chatv1.MarkThreadReadResponse, error) {
	if err := s.ThreadReader.MarkThreadRead(ctx, threaduc.MarkThreadReadInput{UserID: userIDFrom(ctx), ThreadID: req.ThreadId}); err != nil {
		return nil, err
	}
	return &chatv1.MarkThreadReadResponse{}, nil
}

func (s *ThreadServer) FollowThread(ctx context.Context, req *chatv1.FollowThreadRequest) (*chatv1.FollowThreadResponse, error) {
	if err := s.ThreadReader.FollowThread(ctx, threaduc.FollowThreadInput{UserID: userIDFrom(ctx), ThreadID: req.MessageId}); err != nil {
		return nil, err
	}
	return &chatv1.FollowThreadResponse{}, nil
}

func (s *ThreadServer) UnfollowThread(ctx context.Context, req *chatv1.UnfollowThreadRequest) (*chatv1.UnfollowThreadResponse, error) {
	if err := s.ThreadReader.UnfollowThread(ctx, threaduc.FollowThreadInput{UserID: userIDFrom(ctx), ThreadID: req.MessageId}); err != nil {
		return nil, err
	}
	return &chatv1.UnfollowThreadResponse{}, nil
}
