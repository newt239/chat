package datamigration

import "testing"

func TestRewriteLegacyBody(t *testing.T) {
	lc := legacyContext{
		users:    []named{{id: "u-bob", name: "Bob Smith"}},
		groups:   []named{{id: "g-dev", name: "developers"}},
		channels: map[string]string{"general": "c-general", "dev/frontend": "c-front"},
	}
	tests := []struct {
		name        string
		body        string
		want        string
		wantChannel bool
		wantHere    bool
	}{
		{name: "ユーザーは表示名の前方一致", body: "@bob 見て", want: "<@u-bob> 見て"},
		{name: "グループは完全一致", body: "（@developers）", want: "（<@&g-dev>）"},
		{name: "保存されていない名前はそのまま", body: "@alice", want: "@alice"},
		{name: "@channel と @here", body: "@channel と @here", want: "<@channel> と <@here>", wantChannel: true, wantHere: true},
		{name: "チャンネル", body: "#dev/frontend と #general", want: "<#c-front> と <#c-general>"},
		{name: "メールアドレスや URL の途中は書き換えない", body: "bob@bob.com https://example.com/#general", want: "bob@bob.com https://example.com/#general"},
		{name: "コードの中は書き換えない", body: "`@bob` @bob", want: "`@bob` <@u-bob>"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotChannel, gotHere := rewriteLegacyBody(tt.body, lc)
			if got != tt.want || gotChannel != tt.wantChannel || gotHere != tt.wantHere {
				t.Fatalf("got=%q channel=%t here=%t", got, gotChannel, gotHere)
			}
		})
	}
}
