import { useState } from "react";

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
  onOpenList: (emoji: string) => void;
};

// これを超える種類は「+N」にまとめる
const VISIBLE_LIMIT = 10;

export const ReactionList = ({ messageId, reactions, onOpenList }: ReactionListProps) => {
  const { t } = useTranslation();
  const user = useAtomValue(userAtom);
  const toggleReaction = useToggleReaction(messageId);
  const [isExpanded, setIsExpanded] = useState(false);
  const groups = groupReactions(reactions, user?.id ?? null);

  if (groups.length === 0) {
    return null;
  }
  const hiddenCount = groups.length - VISIBLE_LIMIT;
  const visibleGroups = isExpanded ? groups : groups.slice(0, VISIBLE_LIMIT);

  return (
    <div className="flex flex-wrap gap-1">
      {visibleGroups.map((group) => (
        <ReactionButton
          key={group.emoji}
          group={group}
          onPress={() => {
            toggleReaction(group.emoji, group.hasUserReacted);
          }}
          onOpenList={() => {
            onOpenList(group.emoji);
          }}
        />
      ))}
      {hiddenCount > 0 && (
        <Button
          aria-label={isExpanded ? undefined : t("reaction.moreLabel")}
          aria-expanded={isExpanded}
          onPress={() => {
            setIsExpanded(!isExpanded);
          }}
          className={cn(reactionPillClassName, focusRing)}
        >
          <span className="text-xs">
            {isExpanded ? t("reaction.collapse") : t("reaction.more", { count: hiddenCount })}
          </span>
        </Button>
      )}
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
