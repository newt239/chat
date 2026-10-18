import { IconAt, IconBookmark, IconFilePencil, IconMessages } from "@tabler/icons-react";

// チャンネル一覧の上に置く固定の移動先。デスクトップはタブ、モバイルは一覧の行で出す
export const navShortcuts = [
  { icon: IconMessages, label: "shell.nav.threads", to: "/app/$workspaceId/threads" },
  { icon: IconAt, label: "shell.nav.mentions", to: "/app/$workspaceId/mentions" },
  { icon: IconBookmark, label: "shell.nav.bookmarks", to: "/app/$workspaceId/bookmarks" },
  { icon: IconFilePencil, label: "draft.page.title", to: "/app/$workspaceId/drafts" },
] as const;
