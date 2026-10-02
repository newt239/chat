// スレッド一覧・メンション一覧（一覧から返信できる画面）
export const inbox = {
  mention: {
    emptyDescription: "あなた宛てのメンションがここに並びます",
    emptyTitle: "メンションはありません",
    failed: "メンションを読み込めませんでした",
    replyPlaceholder: "{{name}} さんにスレッドで返信…",
  },
  thread: {
    open: "スレッドを開く",
    replyPlaceholder: "返信する…",
    showMore: "他 {{count}} 件の返信を表示",
    unread: "未読 {{count}}",
  },
} as const;
