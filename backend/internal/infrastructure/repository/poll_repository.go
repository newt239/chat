package repository

import (
	"context"
	"time"

	"github.com/newt239/chat/ent"
	"github.com/newt239/chat/ent/poll"
	"github.com/newt239/chat/ent/polloption"
	"github.com/newt239/chat/ent/pollvote"
	"github.com/newt239/chat/internal/domain/entity"
	domainrepository "github.com/newt239/chat/internal/domain/repository"
	"github.com/newt239/chat/internal/infrastructure/transaction"
	"github.com/newt239/chat/internal/infrastructure/utils"
)

type pollRepository struct {
	client *ent.Client
}

func NewPollRepository(client *ent.Client) domainrepository.PollRepository {
	return &pollRepository{client: client}
}

func (r *pollRepository) Create(ctx context.Context, p *entity.Poll) error {
	messageID, err := utils.ParseUUID(p.MessageID, "message ID")
	if err != nil {
		return err
	}
	client := transaction.ResolveClient(ctx, r.client)
	created, err := client.Poll.Create().
		SetMessageID(messageID).
		SetQuestion(p.Question).
		SetMode(poll.Mode(p.Mode)).
		SetAllowMultiple(p.AllowMultiple).
		SetAnonymous(p.Anonymous).
		SetNillableClosesAt(p.ClosesAt).
		Save(ctx)
	if err != nil {
		return err
	}
	builders := make([]*ent.PollOptionCreate, len(p.Options))
	for i, o := range p.Options {
		builders[i] = client.PollOption.Create().
			SetPollID(created.ID).
			SetPosition(i).
			SetLabel(o.Label).
			SetNillableStartsAt(o.StartsAt).
			SetAllDay(o.AllDay)
	}
	options, err := client.PollOption.CreateBulk(builders...).Save(ctx)
	if err != nil {
		return err
	}
	p.ID = created.ID.String()
	p.CreatedAt = created.CreatedAt
	for i, o := range options {
		p.Options[i].ID = o.ID.String()
	}
	return nil
}

func (r *pollRepository) query(ctx context.Context) *ent.PollQuery {
	return transaction.ResolveClient(ctx, r.client).Poll.Query().
		WithOptions(func(q *ent.PollOptionQuery) { q.Order(ent.Asc(polloption.FieldPosition)) })
}

func (r *pollRepository) FindByID(ctx context.Context, id string) (*entity.Poll, error) {
	pid, err := utils.ParseUUID(id, "poll ID")
	if err != nil {
		return nil, err
	}
	found, err := r.query(ctx).Where(poll.ID(pid)).Only(ctx)
	if ent.IsNotFound(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return pollToEntity(found), nil
}

func (r *pollRepository) FindByMessageIDs(ctx context.Context, messageIDs []string) (map[string]*entity.Poll, error) {
	result := map[string]*entity.Poll{}
	if len(messageIDs) == 0 {
		return result, nil
	}
	ids, err := utils.ParseUUIDs(messageIDs, "message ID")
	if err != nil {
		return nil, err
	}
	found, err := r.query(ctx).Where(poll.MessageIDIn(ids...)).All(ctx)
	if err != nil {
		return nil, err
	}
	for _, p := range found {
		result[p.MessageID.String()] = pollToEntity(p)
	}
	return result, nil
}

func (r *pollRepository) FindVotesByPollIDs(ctx context.Context, pollIDs []string) ([]*entity.PollVote, error) {
	if len(pollIDs) == 0 {
		return []*entity.PollVote{}, nil
	}
	ids, err := utils.ParseUUIDs(pollIDs, "poll ID")
	if err != nil {
		return nil, err
	}
	found, err := transaction.ResolveClient(ctx, r.client).PollVote.Query().
		Where(pollvote.HasOptionWith(polloption.PollIDIn(ids...))).
		WithOption().
		Order(ent.Asc(pollvote.FieldCreatedAt)).
		All(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]*entity.PollVote, 0, len(found))
	for _, v := range found {
		result = append(result, &entity.PollVote{PollID: v.Edges.Option.PollID.String(), OptionID: v.OptionID.String(), UserID: v.UserID.String()})
	}
	return result, nil
}

func (r *pollRepository) ReplaceVotes(ctx context.Context, pollID, userID string, optionIDs []string) error {
	pid, err := utils.ParseUUID(pollID, "poll ID")
	if err != nil {
		return err
	}
	uid, err := utils.ParseUUID(userID, "user ID")
	if err != nil {
		return err
	}
	options, err := utils.ParseUUIDs(optionIDs, "option ID")
	if err != nil {
		return err
	}
	client := transaction.ResolveClient(ctx, r.client)
	if _, err := client.PollVote.Delete().
		Where(pollvote.UserID(uid), pollvote.HasOptionWith(polloption.PollID(pid))).
		Exec(ctx); err != nil {
		return err
	}
	builders := make([]*ent.PollVoteCreate, len(options))
	for i, optionID := range options {
		builders[i] = client.PollVote.Create().SetOptionID(optionID).SetUserID(uid)
	}
	return client.PollVote.CreateBulk(builders...).Exec(ctx)
}

func (r *pollRepository) Close(ctx context.Context, id string, closedAt time.Time) error {
	pid, err := utils.ParseUUID(id, "poll ID")
	if err != nil {
		return err
	}
	return transaction.ResolveClient(ctx, r.client).Poll.UpdateOneID(pid).SetClosedAt(closedAt).Exec(ctx)
}

func pollToEntity(p *ent.Poll) *entity.Poll {
	result := &entity.Poll{
		ID:            p.ID.String(),
		MessageID:     p.MessageID.String(),
		Question:      p.Question,
		Mode:          entity.PollMode(p.Mode),
		AllowMultiple: p.AllowMultiple,
		Anonymous:     p.Anonymous,
		ClosesAt:      p.ClosesAt,
		ClosedAt:      p.ClosedAt,
		CreatedAt:     p.CreatedAt,
	}
	for _, o := range p.Edges.Options {
		result.Options = append(result.Options, entity.PollOption{ID: o.ID.String(), Label: o.Label, StartsAt: o.StartsAt, AllDay: o.AllDay})
	}
	return result
}
