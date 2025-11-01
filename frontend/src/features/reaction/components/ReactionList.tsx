import { useMemo } from "react";

import { Group } from "@mantine/core";
import { useAtomValue } from "jotai";

import { useAddReaction, useRemoveReaction } from "../hooks/useReactions";

import AddAnotherEmojiButton from "./AddAnotherEmojiButton";
import { ReactionButton } from "./ReactionButton";

import type { Reaction, ReactionGroup } from "../types";

import { userAtom } from "@/providers/store/auth";

type ReactionListProps = {
  messageId: string;
  reactions: Reaction[];
};

export const ReactionList = ({ messageId, reactions }: ReactionListProps) => {
  const addReaction = useAddReaction();
  const removeReaction = useRemoveReaction();
  const user = useAtomValue(userAtom);

  // リアクションをグループ化
  const reactionGroups = useMemo((): ReactionGroup[] => {
    if (!reactions) return [];

    const groups = new Map<string, ReactionGroup>();

    for (const reaction of reactions) {
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
  }, [reactions, user]);

  const handleReactionClick = async (emoji: string, hasUserReacted: boolean) => {
    if (hasUserReacted) {
      await removeReaction.mutateAsync({ messageId, emoji });
    } else {
      await addReaction.mutateAsync({ messageId, emoji });
    }
  };

  const handleAddReaction = async (emoji: string) => {
    await addReaction.mutateAsync({ messageId, emoji });
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
          users={group.users}
          isActive={group.hasUserReacted}
          onClick={() => handleReactionClick(group.emoji, group.hasUserReacted)}
        />
      ))}
      <AddAnotherEmojiButton onClick={handleAddReaction} />
    </Group>
  );
};
