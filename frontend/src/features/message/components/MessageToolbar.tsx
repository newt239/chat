import {
  IconBookmark,
  IconBookmarkFilled,
  IconDots,
  IconMessageReply,
  IconMoodPlus,
} from "@tabler/icons-react";
import { Separator, Toolbar } from "react-aria-components";
import { useTranslation } from "react-i18next";

import { IconButton } from "#/components/ui/IconButton/IconButton";
import { Menu } from "#/components/ui/Menu/Menu";
import { MenuItem } from "#/components/ui/MenuItem/MenuItem";
import { EmojiPickerPopover } from "#/features/reaction/components/EmojiPickerPopover";

import { quickReactions } from "../utils/quickReactions";

import type { MessageMenuAction } from "../hooks/useMessageMenuActions";

type MessageToolbarProps = {
  actions: MessageMenuAction[];
  isBookmarked: boolean;
  onToggleBookmark: () => void;
  onReplyInThread: () => void;
  onReact: (emoji: string) => void;
  // メニューやピッカーを開いている間はホバーが外れてもツールバーを残す
  onOverlayOpenChange: (isOpen: boolean) => void;
};

const buttonClassName = "size-7 [&_svg]:size-4";

// コードブロックのコピーボタンと重ならないよう、メッセージの上端より上に出す
export const MessageToolbar = ({
  actions,
  isBookmarked,
  onToggleBookmark,
  onReplyInThread,
  onReact,
  onOverlayOpenChange,
}: MessageToolbarProps) => {
  const { t } = useTranslation();

  return (
    <Toolbar
      aria-label={t("message.actions.toolbar")}
      className="absolute -top-7 right-4.5 z-3 flex gap-px rounded-md border border-border bg-raised p-0.5 shadow-md"
    >
      {quickReactions.slice(0, 3).map((emoji) => (
        <IconButton
          key={emoji}
          label={t("message.actions.reactWith", { emoji })}
          className={buttonClassName}
          onPress={() => {
            onReact(emoji);
          }}
        >
          <span className="text-sm leading-none">{emoji}</span>
        </IconButton>
      ))}
      <EmojiPickerPopover
        trigger={
          <IconButton label={t("message.actions.addReaction")} className={buttonClassName}>
            <IconMoodPlus />
          </IconButton>
        }
        onSelect={onReact}
        onOpenChange={onOverlayOpenChange}
      />
      <Separator orientation="vertical" className="mx-0.5 my-1 w-px bg-border" />
      <IconButton
        label={t("message.actions.replyInThread")}
        className={buttonClassName}
        onPress={onReplyInThread}
      >
        <IconMessageReply />
      </IconButton>
      <IconButton
        label={t(isBookmarked ? "message.actions.unbookmark" : "message.actions.bookmark")}
        className={buttonClassName}
        onPress={onToggleBookmark}
      >
        {isBookmarked ? <IconBookmarkFilled className="text-accent-text" /> : <IconBookmark />}
      </IconButton>
      <Menu
        onOpenChange={onOverlayOpenChange}
        trigger={
          <IconButton label={t("message.actions.more")} className={buttonClassName}>
            <IconDots />
          </IconButton>
        }
      >
        {actions.map(({ id, label, icon: ActionIcon, tone, href, onAction }) => (
          <MenuItem
            key={id}
            id={id}
            icon={<ActionIcon />}
            tone={tone}
            href={href}
            target={href === undefined ? undefined : "_blank"}
            onAction={onAction}
          >
            {label}
          </MenuItem>
        ))}
      </Menu>
    </Toolbar>
  );
};
