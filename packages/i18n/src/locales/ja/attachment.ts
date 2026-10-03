export const attachment = {
  completed: "完了",
  crop: {
    tall: "縦長 · 全体を表示",
    wide: "横長 · 全体を表示",
  },
  download: "ダウンロード",
  downloadFailed: "ダウンロードできませんでした",
  errors: {
    empty: "ファイルが空です",
    tooLarge: "ファイルサイズが上限（1GB）を超えています: {{size}}",
    unknown: "アップロードに失敗しました",
  },
  expand: "{{name}} を拡大",
  failed: "エラー: {{error}}",
  lightbox: {
    label: "画像ビューア",
    next: "次の画像",
    page: "{{index}} 枚目",
    position: "{{index}} / {{total}}",
    previous: "前の画像",
    tallHint: "縦長の画像はスクロールで全体を確認できます",
  },
  loadFailed: "読み込めませんでした",
  remove: "添付を外す",
} as const;
