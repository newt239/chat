import { getRouteApi } from "@tanstack/react-router";
import { z } from "zod";

const optionalId = z.string().min(1).optional().catch(undefined);

// ワークスペース内のどの画面にも重ねられる右パネルとダイアログ。子のルートはこれを継承する
export const workspaceSearchSchema = z.object({
  // edit-app で編集するアプリ
  app: optionalId,
  // create-category で作成後にそのカテゴリへ移すチャンネル
  assign: optionalId,
  // edit-category で名前を変えるカテゴリ
  category: optionalId,
  dialog: z
    .enum([
      "create-channel",
      "create-category",
      "edit-category",
      "create-dm",
      "create-group",
      "edit-group",
      "create-workspace",
      "markdown-help",
      "add-link",
      "edit-link",
      "add-app",
      "edit-app",
    ])
    .optional()
    .catch(undefined),
  // リアクション一覧で開いているタブの絵文字。なければすべて
  emoji: optionalId,
  // 右パネルのユーザーグループ
  group: optionalId,
  // ライトボックスで開いている画像の添付
  image: optionalId,
  // edit-link で編集する関連リンク
  link: optionalId,
  // 右パネル。表示中のチャンネルのメンバー・情報・ピン留め
  panel: z.enum(["members", "info", "pins"]).optional().catch(undefined),
  // create-channel で親にするチャンネル
  parent: optionalId,
  // 右パネルのプロフィール
  profile: optionalId,
  // リアクション一覧を開いているメッセージ
  reactions: optionalId,
  // モバイルで長押ししたメッセージの操作シート
  sheet: optionalId,
});

// 右パネルやダイアログの search を子のルートから読む
export const workspaceRoute = getRouteApi("/app/$workspaceId");

type WorkspaceSearch = z.input<typeof workspaceSearchSchema>;

export type PanelSearch = Pick<WorkspaceSearch, "group" | "panel" | "profile">;
type DialogSearch = Omit<WorkspaceSearch, keyof PanelSearch>;

const noPanel = { group: undefined, panel: undefined, profile: undefined } satisfies Record<
  keyof PanelSearch,
  undefined
>;
const noDialog = {
  app: undefined,
  assign: undefined,
  category: undefined,
  dialog: undefined,
  emoji: undefined,
  image: undefined,
  link: undefined,
  parent: undefined,
  reactions: undefined,
  sheet: undefined,
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
