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
  invite: {
    addedDirectly: "{{email}} をワークスペースに追加しました",
    email: "メールアドレスで招待",
    failed: "招待できませんでした",
    link: "{{email}} への招待リンク",
    linkOnce:
      "このリンクは今だけ表示されます。コピーして招待する人に共有してください（7 日間有効）。",
    role: "招待するロール",
    submit: "招待",
  },
  members: {
    remove: "{{name}} をワークスペースから外す",
    role: "ロール",
    title: "メンバー（{{count}} 人）",
  },
  settings: {
    adminOnly: "変更できるのは管理者とオーナーだけです",
    delete: "ワークスペースを削除",
    deleteConfirm: "{{name}} を削除しますか？",
    deleteDescription: "チャンネルとメッセージもすべて削除されます。取り消せません。",
    description: "説明",
    isPublic: "公開ワークスペースにする（誰でも参加できます）",
    name: "名前",
    sections: {
      general: "一般",
      members: "メンバー",
    },
    title: "ワークスペースの設定",
  },
} as const;
