import { useMemo } from "react";

import { Group } from "@mantine/core";
import { useAtomValue } from "jotai";

import { useReactions, useAddReaction, useRemoveReaction } from "../hooks/useReactions";

import { ReactionButton } from "./ReactionButton";

import type { ReactionGroup } from "../types";

import { userAtom } from "@/lib/store/auth";

interface ReactionListProps {
  messageId: string;
}

export const ReactionList = ({ messageId }: ReactionListProps) => {
  const { data, isLoading } = useReactions(messageId);
  const addReaction = useAddReaction();
  const removeReaction = useRemoveReaction();
  const user = useAtomValue(userAtom);

  // リアクションをグループ化
  const reactionGroups = useMemo((): ReactionGroup[] => {
    if (!data?.reactions) return [];

    const groups = new Map<string, ReactionGroup>();

    for (const reaction of data.reactions) {
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
  }, [data?.reactions, user]);

  const handleReactionClick = async (emoji: string, hasUserReacted: boolean) => {
    if (hasUserReacted) {
      await removeReaction.mutateAsync({ messageId, emoji });
    } else {
      await addReaction.mutateAsync({ messageId, emoji });
    }
  };

  if (isLoading || reactionGroups.length === 0) {
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
