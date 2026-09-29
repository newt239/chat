export const auth = {
  displayName: "表示名",
  email: "メールアドレス",
  google: {
    loadFailed: "Google ログインを読み込めませんでした",
  },
  invite: {
    hasAccount: "すでにアカウントをお持ちの方は",
    lead: "{{email}} 宛ての招待です。招待されたメールアドレスの Google アカウントで参加してください。",
    linkTitle: "招待リンク",
    notFound: "招待が見つからないか、有効期限が切れています",
    passwordLead: "Google アカウントを使わない場合は、パスワードを設定して参加できます。",
    submit: "パスワードを設定して参加",
    title: "{{workspace}} への招待",
  },
  login: {
    invitationOnly: "アカウントは管理者からの招待で作成されます",
    or: "または",
    submit: "ログイン",
    title: "ログイン",
  },
  password: "パスワード",
  passwordRule: "8 文字以上",
} as const;
