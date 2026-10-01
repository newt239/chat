import { z } from "zod";

const optionalId = z.string().min(1).optional().catch(undefined);

// ワークスペース内のどの画面にも重ねられる右パネルとダイアログ。子のルートはこれを継承する
export const workspaceSearchSchema = z.object({
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
      "add-webhook",
      "edit-webhook",
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
  // edit-webhook で編集する Webhook
  webhook: optionalId,
  // モバイルで長押ししたメッセージの操作シート
  sheet: optionalId,
});

export type WorkspaceSearch = z.input<typeof workspaceSearchSchema>;
