export const workspaceSettingsSections = ["general", "members", "emoji"] as const;
export type WorkspaceSettingsSection = (typeof workspaceSettingsSections)[number];

export const isWorkspaceSettingsSection = (value: string): value is WorkspaceSettingsSection =>
  workspaceSettingsSections.some((section) => section === value);
