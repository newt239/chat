import { IconMoodPlus } from "@tabler/icons-react";
import { useAtomValue } from "jotai";
import { Button } from "react-aria-components";
import { useTranslation } from "react-i18next";

import { cn, focusRing } from "#/components/ui/styles";
import { userAtom } from "#/providers/store/auth";

import { useToggleReaction } from "../hooks/useReactions";
import { reactionPillClassName } from "../styles";
import { groupReactions } from "../utils/groupReactions";
import { EmojiPickerPopover } from "./EmojiPickerPopover";
import { ReactionButton } from "./ReactionButton";

import type { Reaction } from "#/gen/chat/v1/message_pb";

type ReactionListProps = {
  messageId: string;
  reactions: Reaction[];
};

export const ReactionList = ({ messageId, reactions }: ReactionListProps) => {
  const { t } = useTranslation();
  const user = useAtomValue(userAtom);
  const toggleReaction = useToggleReaction(messageId);
  const groups = groupReactions(reactions, user?.id ?? null);

  if (groups.length === 0) {
    return null;
  }

  return (
    <div className="flex flex-wrap gap-1">
      {groups.map((group) => (
        <ReactionButton
          key={group.emoji}
          group={group}
          onPress={() => {
            toggleReaction(group.emoji, group.hasUserReacted);
          }}
        />
      ))}
      <EmojiPickerPopover
        trigger={
          <Button
            aria-label={t("reaction.add")}
            className={cn(
              reactionPillClassName,
              focusRing,
              "bg-transparent px-1.5 text-subtle [&_svg]:size-3.5",
            )}
          >
            <IconMoodPlus aria-hidden />
          </Button>
        }
        onSelect={(emoji) => {
          toggleReaction(
            emoji,
            groups.some((group) => group.emoji === emoji && group.hasUserReacted),
          );
        }}
      />
    </div>
  );
};
