export const settingsSections = [
  "account",
  "notifications",
  "theme",
  "display",
  "shortcuts",
] as const;
export type SettingsSection = (typeof settingsSections)[number];

// URL の値が設定の画面のどれかなら、その名前を返す
export const findSettingsSection = (value: string) =>
  settingsSections.find((section) => section === value);
