export const schedule = {
  dialog: {
    confirm: "予約する",
    field: "送信日時",
    past: "現在より後の日時を指定してください",
    title: "送信日時を指定",
  },
  failed: "予約できませんでした",
  menu: {
    custom: "日時を指定…",
    label: "送信を予約",
  },
  presets: {
    inOneHour: "1 時間後",
    nextMonday: "次の月曜 9:00",
    tomorrowMorning: "明日の朝 9:00",
  },
  scheduled: "{{time}} に送信を予約しました",
} as const;
