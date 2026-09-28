export const workspace = {
  create: {
    description: "説明（任意）",
    id: "ワークスペース ID",
    idDescription: "URL に使います。小文字の英数字とハイフンで 3〜12 文字",
    idInvalid:
      "小文字の英数字とハイフンで 3〜12 文字にしてください（先頭と末尾にハイフンは使えません）",
    name: "ワークスペース名",
    submit: "作成",
    title: "ワークスペースを作成",
  },
  list: {
    join: "参加する",
    loadFailed: "ワークスペースを読み込めませんでした",
    memberCount: "{{count}} 人",
    open: "開く",
    public: "参加できる公開ワークスペース",
    title: "ワークスペース",
  },
  members: {
    invite: "メールアドレスで招待",
    inviteSubmit: "招待",
    remove: "{{name}} をワークスペースから外す",
    role: "ロール",
    title: "メンバー（{{count}} 人）",
  },
  role: {
    admin: "管理者",
    guest: "ゲスト",
    member: "メンバー",
    owner: "オーナー",
  },
  settings: {
    delete: "ワークスペースを削除",
    deleteConfirm: "{{name}} を削除しますか？",
    deleteDescription: "チャンネルとメッセージもすべて削除されます。取り消せません。",
    description: "説明",
    isPublic: "公開ワークスペースにする（誰でも参加できます）",
    name: "名前",
    title: "ワークスペースの設定",
  },
} as const;
