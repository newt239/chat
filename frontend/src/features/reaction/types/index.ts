export type UserInfo = {
  id: string;
  displayName: string;
  avatarUrl?: string | null;
};

export type ReactionGroup = {
  emoji: string;
  count: number;
  users: UserInfo[]; // ユーザー情報の配列
  hasUserReacted: boolean;
};
