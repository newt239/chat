import { useMemo } from "react";

import { Group } from "@mantine/core";
import { useAtomValue } from "jotai";

import { userAtom } from "#/providers/store/auth";

import { useAddReaction, useRemoveReaction } from "../hooks/useReactions";
import AddAnotherEmojiButton from "./AddAnotherEmojiButton";
import { ReactionButton } from "./ReactionButton";

import type { Reaction, ReactionGroup } from "../types";

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
          count: 1,
          emoji: reaction.emoji,
          hasUserReacted: user ? reaction.user.id === user.id : false,
          users: [reaction.user],
        });
      }
    }

    return [...groups.values()];
  }, [reactions, user]);

  const handleReactionClick = async (emoji: string, hasUserReacted: boolean) => {
    await (hasUserReacted ? removeReaction : addReaction).mutateAsync({ emoji, messageId });
  };

  const handleAddReaction = async (emoji: string) => {
    await addReaction.mutateAsync({ emoji, messageId });
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
          onClick={() => {
            void handleReactionClick(group.emoji, group.hasUserReacted);
          }}
        />
      ))}
      <AddAnotherEmojiButton
        onClick={(emoji) => {
          void handleAddReaction(emoji);
        }}
      />
    </Group>
  );
};
