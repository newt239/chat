import { useEffect, useState } from "react";

import { useAtomValue } from "jotai";
import { motion } from "motion/react";
import { Button } from "react-aria-components";
import { useTranslation } from "react-i18next";

import { cn, focusRing } from "#/components/ui/styles";
import { Tooltip } from "#/components/ui/Tooltip";
import { transitions } from "#/lib/motion";
import { preferencesAtom } from "#/providers/store/preferences";

import { reactionPillClassName } from "../styles";

import type { ReactionGroup } from "../types";

type ReactionButtonProps = {
  group: ReactionGroup;
  onPress: () => void;
};

export const ReactionButton = ({ group, onPress }: ReactionButtonProps) => {
  const { t } = useTranslation();
  // 初回の描画では弾ませず、件数が変わったときだけ弾ませる
  const [isMounted, setIsMounted] = useState(false);
  useEffect(() => {
    setIsMounted(true);
  }, []);
  const { locale } = useAtomValue(preferencesAtom);
  const names = new Intl.ListFormat(locale).format(group.users.map((user) => user.displayName));

  return (
    <Tooltip content={names}>
      <Button
        aria-label={t("reaction.summary", { count: group.count, emoji: group.emoji, names })}
        aria-pressed={group.hasUserReacted}
        onPress={onPress}
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
          {group.emoji}
        </motion.span>
        <span className="text-xs">{group.count}</span>
      </Button>
    </Tooltip>
  );
};
