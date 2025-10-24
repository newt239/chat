import type { components } from "@/lib/api/schema";

export type ReactionWithUser = components["schemas"]["ReactionWithUser"];

export type UserInfo = {
  id: string;
  displayName: string;
  avatarUrl?: string | null;
}

export type ReactionGroup = {
  emoji: string;
  count: number;
  users: UserInfo[]; // ユーザー情報の配列
  hasUserReacted: boolean;
}

// 将来的なカスタム絵文字対応のための型
export type EmojiData = {
  id: string; // Unicode絵文字の場合は絵文字自体、カスタムの場合はID
  native?: string; // Unicode絵文字
  imageUrl?: string; // カスタム絵文字の画像URL
  name: string; // 絵文字の名前
  isCustom: boolean; // カスタム絵文字かどうか
}
