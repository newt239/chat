import type { Reaction } from "#/gen/chat/v1/message_pb";
import type { UserSummary } from "#/gen/chat/v1/user_pb";

// リアクションの一覧で、絵文字ごとではなくすべてを並べるタブ
export const ALL_REACTIONS_TAB = "all";

export type ReactionGroup = {
  emoji: string;
  count: number;
  users: UserSummary[];
  hasUserReacted: boolean;
};

// 絵文字ごとにまとめる。並びは最初に付いた順
export const groupReactions = (reactions: Reaction[], currentUserId: string | null) => {
  const groups = new Map<string, ReactionGroup>();
  for (const { emoji, user } of reactions) {
    const group = groups.get(emoji) ?? { count: 0, emoji, hasUserReacted: false, users: [] };
    group.count++;
    if (user !== undefined) {
      group.users.push(user);
      group.hasUserReacted ||= user.id === currentUserId;
    }
    groups.set(emoji, group);
  }
  return [...groups.values()];
};
