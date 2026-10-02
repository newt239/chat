package repository

import (
	"context"
	"testing"
	"time"

	"github.com/newt239/chat/internal/domain/entity"
)

func TestPollRepositoryVotes(t *testing.T) {
	client := openTestClient(t)
	f := newSearchFixture(t, client)
	repo := NewPollRepository(client)
	ctx := context.Background()
	day := time.Date(2026, 10, 10, 0, 0, 0, 0, time.UTC)

	poll := &entity.Poll{
		MessageID: f.messages["mention"].ID.String(),
		Question:  "打ち上げ",
		Mode:      entity.PollModeDate,
		Options:   []entity.PollOption{{StartsAt: &day, AllDay: true}, {StartsAt: &day}},
	}
	if err := repo.Create(ctx, poll); err != nil {
		t.Fatal(err)
	}
	a, b := poll.Options[0].ID, poll.Options[1].ID
	alice, bob := f.alice.ID.String(), f.bob.ID.String()
	for _, step := range []struct {
		user    string
		options []string
	}{{alice, []string{a, b}}, {bob, []string{b}}, {alice, []string{b}}} {
		if err := repo.ReplaceVotes(ctx, poll.ID, step.user, step.options); err != nil {
			t.Fatal(err)
		}
	}

	votes, err := repo.FindVotesByPollIDs(ctx, []string{poll.ID})
	if err != nil {
		t.Fatal(err)
	}
	if len(votes) != 2 || votes[0].OptionID != b || votes[1].OptionID != b || votes[0].PollID != poll.ID {
		t.Fatalf("票が置き換わっていません: %+v", votes)
	}
	found, err := repo.FindByMessageIDs(ctx, []string{poll.MessageID})
	if err != nil {
		t.Fatal(err)
	}
	if got := found[poll.MessageID]; got == nil || len(got.Options) != 2 || got.Options[0].ID != a || !got.Options[0].AllDay {
		t.Fatalf("選択肢が並び順で読めません: %+v", got)
	}
}
