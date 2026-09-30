import { useEffect, useState } from "react";

import { motion } from "motion/react";
import { Button } from "react-aria-components";
import { useTranslation } from "react-i18next";

import { cn, focusRing } from "#/components/ui/styles/styles";
import { Tooltip } from "#/components/ui/Tooltip/Tooltip";
import { transitions } from "#/lib/motion";

import { useFormatReactors } from "../hooks/useFormatReactors";
import { reactionPillClassName } from "../styles";
import { ReactionEmoji } from "./ReactionEmoji";

import type { ReactionGroup } from "#/features/reaction/types/reactionGroup";

type ReactionButtonProps = {
  group: ReactionGroup;
  onPress: () => void;
  // 右クリックで誰がいつ押したかの一覧を開く
  onOpenList: () => void;
};

export const ReactionButton = ({ group, onPress, onOpenList }: ReactionButtonProps) => {
  const { t } = useTranslation();
  // 初回の描画では弾ませず、件数が変わったときだけ弾ませる
  const [isMounted, setIsMounted] = useState(false);
  useEffect(() => {
    setIsMounted(true);
  }, []);
  const names = useFormatReactors()(group.users);

  return (
    <Tooltip
      content={
        <span className="flex flex-col items-center px-1 py-0.5 text-center leading-normal">
          <span className="text-[28px] leading-[1.2]">
            <ReactionEmoji emoji={group.emoji} />
          </span>
          {t("reaction.tooltip.reacted", { names })}
          <small className="mt-0.5 text-[11px] opacity-60">{t("reaction.tooltip.hint")}</small>
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
        <motion.span
          key={`${group.emoji}-${group.count}`}
          initial={isMounted ? { scale: 0.6 } : false}
          animate={{ scale: 1 }}
          transition={transitions.spring}
          className="text-sm leading-none"
        >
          <ReactionEmoji emoji={group.emoji} />
        </motion.span>
        <span className="text-xs">{group.count}</span>
      </Button>
    </Tooltip>
  );
};
