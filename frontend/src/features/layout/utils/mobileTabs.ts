import { IconBell, IconHome, IconMessageCircle, IconUser } from "@tabler/icons-react";

// モバイルのボトムタブ。to はタブの画面のパス
export const mobileTabs = [
  { icon: IconHome, name: "home", to: "/app/$workspaceId" },
  { icon: IconMessageCircle, name: "dms", to: "/app/$workspaceId/dms" },
  { icon: IconBell, name: "activity", to: "/app/$workspaceId/mentions" },
  { icon: IconUser, name: "me", to: "/app/$workspaceId/me" },
] as const;

export type MobileTab = (typeof mobileTabs)[number];
