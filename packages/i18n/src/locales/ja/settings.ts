export const settings = {
  account: {
    changePassword: "パスワードを変更",
    currentPassword: "現在のパスワード",
    delete: "アカウントを削除",
    deleteConfirm: "アカウントを削除しますか？",
    deleteDescription:
      "投稿したメッセージを含め、アカウントのデータが削除されます。取り消せません。",
    newPassword: "新しいパスワード",
    openProfile: "プロフィールを開く",
    password: "パスワードの変更",
    passwordChanged: "パスワードを変更しました。もう一度ログインしてください",
    profile: "プロフィール",
    profileDescription: "名前や自己紹介はプロフィールで編集します",
  },
  notifications: {
    denied: "ブラウザで通知が許可されていません",
    desktop: "デスクトップ通知",
    desktopDescription: "この端末のブラウザに通知を出します",
    level: "通知するメッセージ",
    levels: {
      all: "すべて",
      mentions: "メンションと DM",
      none: "通知しない",
    },
    muteHint:
      "ミュートしたチャンネルと DM は通知しません。ミュートはチャンネルの見出しの「その他」から設定できます。",
  },
  profile: {
    avatarUrl: "アイコンの URL",
    bio: "自己紹介",
    displayNameDescription: "メッセージやメンションで表示される名前です",
    saved: "プロフィールを保存しました",
  },
  sections: {
    account: "アカウント",
    display: "表示",
    notifications: "通知",
    shortcuts: "ショートカット",
    theme: "テーマ",
  },
  shortcuts: {
    close: "パネル・ダイアログを閉じる",
    newTab: "新しいタブで開く",
    newline: "改行",
    search: "検索",
    send: "送信",
    settings: "設定を開く",
  },
  title: "設定",
} as const;
