import { useState } from "react";

import { IconMoodPlus } from "@tabler/icons-react";
import { Button, Heading, Menu } from "react-aria-components";
import { useTranslation } from "react-i18next";

import { DialogFrame } from "#/components/ui/Dialog/DialogFrame";
import { MenuItem } from "#/components/ui/MenuItem/MenuItem";
import { MenuSeparator } from "#/components/ui/MenuSeparator/MenuSeparator";
import { focusRing } from "#/components/ui/styles/styles";
import { useDisplayName } from "#/features/member/hooks/useDisplayName";
import { useMentionDirectory } from "#/features/mention/hooks/useMentionDirectory";
import { EmojiPicker } from "#/features/reaction/components/EmojiPicker";

import { quickReactions } from "../hooks/useMessageMenuActions";

import type { MessageMenuAction } from "../hooks/useMessageMenuActions";

import type { Message } from "#/gen/chat/v1/message_pb";

type MessageActionSheetProps = {
  onClose: () => void;
  message: Message;
  actions: MessageMenuAction[];
  onReact: (emoji: string) => void;
};

const roundButtonClassName = `grid size-11 place-items-center rounded-full bg-sunken text-heading font-normal text-muted data-pressed:bg-hover [&_svg]:size-5.25 ${focusRing}`;

// モバイルでメッセージを長押ししたときの操作。開いている間だけマウントする。新しいタブで開く操作はモバイルでは出さない
export const MessageActionSheet = ({
  onClose,
  message,
  actions,
  onReact,
}: MessageActionSheetProps) => {
  const { t } = useTranslation();
  const { toExcerpt } = useMentionDirectory();
  const [isPickingEmoji, setIsPickingEmoji] = useState(false);
  const displayName = useDisplayName();
  const react = (emoji: string) => {
    onReact(emoji);
    onClose();
  };

  return (
    <DialogFrame
      isOpen
      onOpenChange={onClose}
      layout="bottom"
      role="dialog"
      className="bg-raised pt-2"
    >
      <Heading slot="title" className="sr-only">
        {t("message.sheet.title")}
      </Heading>
      {isPickingEmoji ? (
        <div className="flex justify-center px-3">
          <EmojiPicker onEmojiSelect={react} />
        </div>
      ) : (
        <>
          <div className="mx-4 mb-3 flex flex-col gap-0.5 rounded-lg bg-sunken px-3 py-2.5 text-body-sm text-muted">
            <b className="text-text">
              {displayName(message.userId, message.user?.displayName ?? "")}
            </b>
            <span className="line-clamp-2">{toExcerpt(message.body)}</span>
          </div>
          <div className="flex justify-between px-4 pb-2.5">
            {quickReactions.map((emoji) => (
              <Button
                key={emoji}
                aria-label={t("message.actions.reactWith", { emoji })}
                className={roundButtonClassName}
                onPress={() => {
                  react(emoji);
                }}
              >
                <span className="text-heading font-normal leading-none">{emoji}</span>
              </Button>
            ))}
            <Button
              aria-label={t("message.actions.moreEmoji")}
              className={roundButtonClassName}
              onPress={() => {
                setIsPickingEmoji(true);
              }}
            >
              <IconMoodPlus aria-hidden />
            </Button>
          </div>
          <Menu aria-label={t("message.sheet.title")} className="flex flex-col px-3 outline-none">
            {actions
              .filter(({ href }) => href === undefined)
              .flatMap(({ id, label, icon: ActionIcon, tone, onAction }) => [
                tone === "danger" && <MenuSeparator key={`${id}-separator`} />,
                <MenuItem
                  key={id}
                  id={id}
                  icon={<ActionIcon aria-hidden />}
                  tone={tone}
                  onAction={() => {
                    onClose();
                    onAction?.();
                  }}
                >
                  {label}
                </MenuItem>,
              ])}
          </Menu>
        </>
      )}
    </DialogFrame>
  );
};
