import { useMemo } from "react";

import { Group } from "@mantine/core";
import { useAtomValue } from "jotai";

import { userAtom } from "#/providers/store/auth";

import { useAddReaction, useRemoveReaction } from "../hooks/useReactions";
import AddAnotherEmojiButton from "./AddAnotherEmojiButton";
import { ReactionButton } from "./ReactionButton";

import type { ReactionGroup } from "../types";

import type { Reaction } from "#/gen/chat/v1/message_pb";

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

    for (const { emoji, user: reactedUser } of reactions) {
      const users = reactedUser === undefined ? [] : [reactedUser];
      const hasUserReacted = user !== null && reactedUser?.id === user.id;
      const existing = groups.get(emoji);
      if (existing) {
        existing.count++;
        existing.users.push(...users);
        existing.hasUserReacted ||= hasUserReacted;
      } else {
        groups.set(emoji, { count: 1, emoji, hasUserReacted, users });
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
