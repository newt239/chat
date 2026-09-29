type SchedulePreset = "inOneHour" | "tomorrowMorning" | "nextMonday";

const atNine = (date: Date, daysLater: number) =>
  new Date(date.getFullYear(), date.getMonth(), date.getDate() + daysLater, 9, 0);

// 送信メニューの選択肢。月曜は今日が月曜でも翌週にする
export const schedulePresets = (now: Date) => {
  const inOneHour = new Date(now.getTime() + 60 * 60 * 1000);
  inOneHour.setSeconds(0, 0);
  const daysToMonday = (8 - now.getDay()) % 7 || 7;
  const presets: { key: SchedulePreset; date: Date }[] = [
    { date: inOneHour, key: "inOneHour" },
    { date: atNine(now, 1), key: "tomorrowMorning" },
    { date: atNine(now, daysToMonday), key: "nextMonday" },
  ];
  return presets;
};

// 日時を指定するときの初期値
export const defaultScheduleDate = (now: Date) => atNine(now, 1);
