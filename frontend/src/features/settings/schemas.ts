export const settingsSections = [
  "account",
  "notifications",
  "theme",
  "display",
  "shortcuts",
] as const;
export type SettingsSection = (typeof settingsSections)[number];

export const isSettingsSection = (value: string): value is SettingsSection =>
  settingsSections.some((section) => section === value);
