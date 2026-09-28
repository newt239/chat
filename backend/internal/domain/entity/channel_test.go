package entity

import (
	"errors"
	"slices"
	"strings"
	"testing"

	domerr "github.com/newt239/chat/internal/domain/errors"
)

func TestNormalizeChannelPath(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{name: "単一階層", input: "general", want: "general"},
		{name: "英字は小文字にそろえる", input: " Dev/Frontend ", want: "dev/frontend"},
		{name: "ハイフン・アンダースコア・数字", input: "dev2/front-end_web", want: "dev2/front-end_web"},
		{name: "日本語は不可", input: "開発/web", wantErr: true},
		{name: "4 階層まで", input: "a/b/c/d", want: "a/b/c/d"},
		{name: "5 階層は不可", input: "a/b/c/d/e", wantErr: true},
		{name: "空のセグメントは不可", input: "dev//web", wantErr: true},
		{name: "末尾のスラッシュは不可", input: "dev/", wantErr: true},
		{name: "空白を含むセグメントは不可", input: "dev team", wantErr: true},
		{name: "記号は不可", input: "dev.web", wantErr: true},
		{name: "空文字は不可", input: "  ", wantErr: true},
		{name: "32 文字を超えるセグメントは不可", input: strings.Repeat("a", MaxChannelSegmentLength+1), wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NormalizeChannelPath(tt.input)
			if tt.wantErr {
				if !errors.Is(err, domerr.ErrValidation) {
					t.Fatalf("検証エラーになっていません: got=%q err=%v", got, err)
				}
				return
			}
			if err != nil || got != tt.want {
				t.Fatalf("got=%q err=%v, want=%q", got, err, tt.want)
			}
		})
	}
}

func TestAncestorChannelPaths(t *testing.T) {
	if got := AncestorChannelPaths("dev/frontend/web"); !slices.Equal(got, []string{"dev", "dev/frontend"}) {
		t.Fatalf("祖先のパスが正しくありません: %v", got)
	}
	if got := AncestorChannelPaths("general"); len(got) != 0 {
		t.Fatalf("最上位のチャンネルに祖先があります: %v", got)
	}
}

func TestChannelChangeNameKeepsParent(t *testing.T) {
	ch := &Channel{Name: "dev/frontend"}

	if err := ch.ChangeName("dev/web"); err != nil || ch.Name != "dev/web" {
		t.Fatalf("末尾の階層を変更できません: name=%q err=%v", ch.Name, err)
	}
	if err := ch.ChangeName("ops/web"); !errors.Is(err, domerr.ErrValidation) {
		t.Fatalf("親の階層を変える変更が拒否されていません: %v", err)
	}
}
