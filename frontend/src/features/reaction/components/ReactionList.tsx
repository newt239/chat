import { useState } from "react";

import { IconMoodPlus } from "@tabler/icons-react";
import { useAtomValue } from "jotai";
import { Button } from "react-aria-components";
import { useTranslation } from "react-i18next";

import { cn, focusRing } from "#/components/ui/styles/styles";
import { myUserIdAtom } from "#/providers/store/auth";

import { groupReactions } from "../utils/groupReactions";
import { reactionPillClassName } from "../utils/reactionPillClassName";
import { EmojiPickerPopover } from "./EmojiPickerPopover";
import { ReactionButton } from "./ReactionButton";

import type { Message } from "#/gen/chat/v1/message_pb";

type ReactionListProps = {
  message: Message;
  onOpenList: (emoji: string) => void;
  onToggleReaction: (emoji: string) => void;
};

// これを超える種類は「+N」にまとめる
const VISIBLE_LIMIT = 10;

export const ReactionList = ({ message, onOpenList, onToggleReaction }: ReactionListProps) => {
  const { t } = useTranslation();
  const myId = useAtomValue(myUserIdAtom);
  const [isExpanded, setIsExpanded] = useState(false);
  const groups = groupReactions(message.reactions, myId);

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
            onToggleReaction(group.emoji);
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
        onSelect={onToggleReaction}
        onOpenChange={null}
        label={t("reaction.add")}
        placement="bottom end"
      />
    </div>
  );
};
