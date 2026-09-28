// キーは「機能.文脈.項目」の camelCase。日本語辞書が正で、英語はこの形に合わせる
export const ja = {
  codeBlock: {
    copied: "コピーしました",
    copy: "コピー",
    copyFailed: "コピーできませんでした",
    lines: "{{count}} 行",
  },
  common: {
    cancel: "キャンセル",
    close: "閉じる",
    delete: "削除",
    loading: "読み込み中",
    ok: "OK",
    save: "保存",
  },
  preferences: {
    locale: {
      en: "English",
      ja: "日本語",
      title: "言語",
    },
    mode: {
      dark: "ダーク",
      light: "ライト",
      system: "システムに合わせる",
      title: "表示モード",
    },
    saveFailed: "設定を保存できませんでした",
    theme: {
      chroma: "彩度",
      custom: "カスタム",
      hue: "色相",
      presets: {
        cobalt: "Cobalt",
        graphite: "Graphite",
        jade: "Jade",
        plum: "Plum",
      },
      sidebar: {
        light: "明るい",
        tinted: "色付き",
        title: "サイドバー",
      },
      title: "テーマ",
    },
  },
  ui: {
    avatar: {
      groupMembers: "{{count}} 人のグループ",
    },
    comboBox: {
      empty: "候補がありません",
      showSuggestions: "候補を表示",
    },
    toast: {
      dismiss: "通知を閉じる",
      region: "通知",
    },
  },
} as const;
