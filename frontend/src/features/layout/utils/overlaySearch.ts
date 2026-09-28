import type { WorkspaceSearch } from "../schemas";

export type PanelSearch = Pick<WorkspaceSearch, "group" | "panel" | "profile">;
type DialogSearch = Omit<WorkspaceSearch, keyof PanelSearch>;

const noPanel = { group: undefined, panel: undefined, profile: undefined } satisfies Record<
  keyof PanelSearch,
  undefined
>;
const noDialog = {
  dialog: undefined,
  emoji: undefined,
  image: undefined,
  link: undefined,
  reactions: undefined,
  settings: undefined,
  sheet: undefined,
  webhook: undefined,
} satisfies Record<keyof DialogSearch, undefined>;

// Link / navigate の search に渡す。今のルートの search（?message= など）は残す
// 右パネルはひとつだけ開く。パネルを開くとダイアログも閉じる
export const openPanel =
  (panel: PanelSearch) =>
  <T extends object>(prev: T) => ({ ...prev, ...noPanel, ...noDialog, ...panel });

// ダイアログはひとつだけ開く。右パネルはその下に残す
export const openDialog =
  (dialog: DialogSearch) =>
  <T extends object>(prev: T) => ({ ...prev, ...noDialog, ...dialog });

export const closePanel = <T extends object>(prev: T) => ({ ...prev, ...noPanel });

export const closeDialog = <T extends object>(prev: T) => ({ ...prev, ...noDialog });
