import { AnimatePresence, motion } from "motion/react";
import { Button } from "react-aria-components";
import { useTranslation } from "react-i18next";

import { cn, focusRing } from "#/components/ui/styles/styles";
import { Tooltip } from "#/components/ui/Tooltip/Tooltip";
import { transitions } from "#/lib/motion";

import { useFormatReactors } from "../hooks/useFormatReactors";
import { ReactionEmoji } from "./ReactionEmoji";

import type { ReactionGroup } from "../utils/groupReactions";

const reactionPillClassName =
  "inline-flex h-6 cursor-pointer items-center gap-1 rounded-full border border-border bg-sunken px-1.75 font-sans text-xs text-muted tabular-nums data-hovered:border-border-strong";

type ReactionButtonProps = {
  group: ReactionGroup;
  onPress: () => void;
  // 右クリックで誰がいつ押したかの一覧を開く
  onOpenList: () => void;
};

export const ReactionButton = ({ group, onPress, onOpenList }: ReactionButtonProps) => {
  const { t } = useTranslation();
  const names = useFormatReactors()(group.users);

  return (
    <Tooltip
      content={
        <span className="flex flex-col items-center px-1 py-0.5 text-center leading-normal">
          <span className="text-emoji leading-tight">
            <ReactionEmoji emoji={group.emoji} />
          </span>
          {t("reaction.tooltip.reacted", { names })}
          <small className="mt-0.5 text-caption opacity-60">{t("reaction.tooltip.hint")}</small>
        </span>
      }
    >
      <Button
        aria-label={t("reaction.summary", { count: group.count, emoji: group.emoji, names })}
        aria-pressed={group.hasUserReacted}
        onPress={onPress}
        onContextMenu={(event) => {
          event.preventDefault();
          onOpenList();
        }}
        className={cn(
          reactionPillClassName,
          focusRing,
          group.hasUserReacted && "border-accent bg-accent-soft font-semibold text-accent-text",
        )}
      >
        {/* 件数が変わると key が変わって弾む。最初に描いたときは弾ませない */}
        <AnimatePresence initial={false}>
          <motion.span
            key={`${group.emoji}-${group.count}`}
            initial={{ scale: 0.6 }}
            animate={{ scale: 1 }}
            transition={transitions.spring}
            className="text-sm leading-none"
          >
            <ReactionEmoji emoji={group.emoji} />
          </motion.span>
        </AnimatePresence>
        <span className="text-xs">{group.count}</span>
      </Button>
    </Tooltip>
  );
};
