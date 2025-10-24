import { useMemo } from "react";

import { Group } from "@mantine/core";
import { useAtomValue } from "jotai";

import { useAddReaction, useRemoveReaction } from "../hooks/useReactions";

import { ReactionButton } from "./ReactionButton";

import type { ReactionGroup } from "../types";
import type { MessageWithUser } from "@/features/message/schemas";

import { userAtom } from "@/lib/store/auth";

type ReactionListProps = {
  message: MessageWithUser;
}

export const ReactionList = ({ message }: ReactionListProps) => {
  const addReaction = useAddReaction();
  const removeReaction = useRemoveReaction();
  const user = useAtomValue(userAtom);

  // リアクションをグループ化
  const reactionGroups = useMemo((): ReactionGroup[] => {
    if (!message.reactions) return [];

    const groups = new Map<string, ReactionGroup>();

    for (const reaction of message.reactions) {
      const existing = groups.get(reaction.emoji);
      if (existing) {
        existing.count++;
        existing.users.push(reaction.user);
        if (user && reaction.user.id === user.id) {
          existing.hasUserReacted = true;
        }
      } else {
        groups.set(reaction.emoji, {
          emoji: reaction.emoji,
          count: 1,
          users: [reaction.user],
          hasUserReacted: user ? reaction.user.id === user.id : false,
        });
      }
    }

    return Array.from(groups.values());
  }, [message.reactions, user]);

  const handleReactionClick = async (emoji: string, hasUserReacted: boolean) => {
    if (hasUserReacted) {
      await removeReaction.mutateAsync({ messageId: message.id, emoji });
    } else {
      await addReaction.mutateAsync({ messageId: message.id, emoji });
    }
  };

  if (reactionGroups.length === 0) {
    return null;
  }

  return (
    <Group gap="xs" mt="xs">
      {reactionGroups.map((group) => (
        <ReactionButton
          key={group.emoji}
          emoji={group.emoji}
          count={group.count}
          users={group.users}
          isActive={group.hasUserReacted}
          onClick={() => handleReactionClick(group.emoji, group.hasUserReacted)}
        />
      ))}
    </Group>
  );
};
