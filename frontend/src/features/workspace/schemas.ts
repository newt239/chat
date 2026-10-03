export const workspaceSettingsSections = ["general", "emoji"] as const;
export type WorkspaceSettingsSection = (typeof workspaceSettingsSections)[number];

// URL の値がワークスペース設定の画面のどれかなら、その名前を返す
export const findWorkspaceSettingsSection = (value: string | undefined) =>
  workspaceSettingsSections.find((section) => section === value);
