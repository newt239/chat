export const preferences = {
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
    customActive: "使用中",
    hue: "色相",
    presetsTitle: "プリセット",
    presets: {
      cobalt: "Cobalt",
      graphite: "Graphite",
      jade: "Jade",
      plum: "Plum",
    },
    preview: "プレビュー",
    sidebar: {
      light: "明るい",
      tinted: "色付き",
      title: "サイドバー",
    },
    title: "テーマ",
  },
  timezone: {
    autoUpdate: "タイムゾーンを自動で更新する",
    autoUpdateDescription: "端末のタイムゾーンが変わったとき、確認せずに更新します",
    changed: "端末のタイムゾーンが {{timezone}} になっています",
    changedDescription: "アカウントのタイムゾーン（{{current}}）を更新しますか？",
    description: "プロフィールに現地時刻として表示されます",
    placeholder: "未設定",
    title: "タイムゾーン",
    update: "更新する",
  },
} as const;
