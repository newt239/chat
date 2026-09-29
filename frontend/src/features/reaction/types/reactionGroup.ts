import type { UserSummary } from "#/gen/chat/v1/user_pb";

export type ReactionGroup = {
  emoji: string;
  count: number;
  users: UserSummary[];
  hasUserReacted: boolean;
};
