import { useState } from "react";

import { IconMoodPlus } from "@tabler/icons-react";
import { Button, Heading, Menu, MenuItem, Separator } from "react-aria-components";
import { useTranslation } from "react-i18next";

import { DialogFrame } from "#/components/ui/Dialog/DialogFrame";
import { cn, focusRing } from "#/components/ui/styles/styles";
import { useDisplayName } from "#/features/member/hooks/useDisplayName";
import { EmojiPicker } from "#/features/reaction/components/EmojiPicker";

import { useMentionDirectory } from "../hooks/useMentionDirectory";
import { toPlainText } from "../utils/markdown/plainText";
import { quickReactions } from "../utils/quickReactions";

import type { MessageMenuAction } from "../hooks/useMessageMenuActions";

import type { Message } from "#/gen/chat/v1/message_pb";

type MessageActionSheetProps = {
  isOpen: boolean;
  onOpenChange: (isOpen: boolean) => void;
  message: Message;
  actions: MessageMenuAction[];
  onReact: (emoji: string) => void;
};

const roundButtonClassName = `grid size-11 place-items-center rounded-full bg-sunken text-[21px] text-muted data-pressed:bg-hover [&_svg]:size-[21px] ${focusRing}`;

// モバイルでメッセージを長押ししたときの操作。新しいタブで開く操作はモバイルでは出さない
export const MessageActionSheet = ({
  isOpen,
  onOpenChange,
  message,
  actions,
  onReact,
}: MessageActionSheetProps) => {
  const { t } = useTranslation();
  const { toText } = useMentionDirectory();
  const [isPickingEmoji, setIsPickingEmoji] = useState(false);
  const displayName = useDisplayName();
  const close = () => {
    onOpenChange(false);
    setIsPickingEmoji(false);
  };
  const react = (emoji: string) => {
    onReact(emoji);
    close();
  };

  return (
    <DialogFrame
      isOpen={isOpen}
      onOpenChange={(next) => {
        if (!next) {
          close();
        }
      }}
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
          <div className="mx-4 mb-3 flex flex-col gap-0.5 rounded-lg bg-sunken px-3 py-2.5 text-[13px] text-muted">
            <b className="text-text">
              {displayName(message.userId, message.user?.displayName ?? "")}
            </b>
            <span className="line-clamp-2">
              {toPlainText(toText(message.body)) || t("message.sheet.attachmentOnly")}
            </span>
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
                <span className="text-[21px] leading-none">{emoji}</span>
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
          <Menu aria-label={t("message.sheet.title")} className="flex flex-col outline-none">
            {actions
              .filter(({ href }) => href === undefined)
              .flatMap(({ id, label, icon: ActionIcon, tone, onAction }) => [
                tone === "danger" && (
                  <Separator key={`${id}-separator`} className="mx-5 my-1 h-px bg-border" />
                ),
                <MenuItem
                  key={id}
                  id={id}
                  textValue={label}
                  onAction={() => {
                    close();
                    onAction?.();
                  }}
                  className={cn(
                    "flex min-h-12 cursor-default items-center gap-3.5 px-5 text-[15.5px] text-text outline-none data-focus-visible:bg-hover data-pressed:bg-hover [&_svg]:size-[21px] [&_svg]:text-muted",
                    tone === "danger" && "text-danger [&_svg]:text-danger",
                  )}
                >
                  <ActionIcon aria-hidden />
                  {label}
                </MenuItem>,
              ])}
          </Menu>
        </>
      )}
    </DialogFrame>
  );
};
