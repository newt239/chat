import { useAtomValue } from "jotai";
import { AnimatePresence, motion } from "motion/react";
import { Button } from "react-aria-components";
import { useTranslation } from "react-i18next";

import { cn, focusRing } from "#/components/ui/styles/styles";
import { Tooltip } from "#/components/ui/Tooltip/Tooltip";
import { useDisplayName } from "#/features/member/hooks/useDisplayName";
import { usePreferences } from "#/hooks/usePreferences";
import { transitions } from "#/lib/motion";
import { myUserIdAtom } from "#/providers/store/auth";

import { reactionPillClassName } from "../utils/reactionPillClassName";
import { ReactionEmoji } from "./ReactionEmoji";

import type { ReactionGroup } from "../utils/groupReactions";

const MAX_NAMES = 4;

type ReactionButtonProps = {
  group: ReactionGroup;
  onPress: () => void;
  // 右クリックで誰がいつ押したかの一覧を開く
  onOpenList: () => void;
};

export const ReactionButton = ({ group, onPress, onOpenList }: ReactionButtonProps) => {
  const { t } = useTranslation();
  const { locale } = usePreferences();
  const currentUserId = useAtomValue(myUserIdAtom);
  const displayName = useDisplayName();
  const listFormat = new Intl.ListFormat(locale);
  // 自分を「あなた」として先頭に置き、多いときは先頭の 3 人と残りの人数にまとめる
  const allNames = group.users
    .toSorted((a, b) => Number(b.id === currentUserId) - Number(a.id === currentUserId))
    .map((user) =>
      user.id === currentUserId ? t("reaction.names.you") : displayName(user.id, user.displayName),
    );
  const names =
    allNames.length <= MAX_NAMES
      ? listFormat.format(allNames)
      : t("reaction.names.others", {
          count: allNames.length - (MAX_NAMES - 1),
          names: listFormat.format(allNames.slice(0, MAX_NAMES - 1)),
        });

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
