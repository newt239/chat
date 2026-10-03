import { useRef } from "react";

import {
  IconArrowUp,
  IconBold,
  IconChartBar,
  IconCode,
  IconEye,
  IconH1,
  IconHelp,
  IconItalic,
  IconLink,
  IconList,
  IconListNumbers,
  IconLoader2,
  IconMapPin,
  IconMicrophone,
  IconMoodSmile,
  IconPaperclip,
  IconQuote,
  IconStrikethrough,
} from "@tabler/icons-react";
import { useNavigate } from "@tanstack/react-router";
import { FileTrigger } from "react-aria-components";
import { useTranslation } from "react-i18next";

import { IconButton } from "#/components/ui/IconButton/IconButton";
import { IconToggleButton } from "#/components/ui/IconToggleButton/IconToggleButton";
import { openDialog } from "#/features/layout/utils/overlaySearch";
import { EmojiPickerPopover } from "#/features/reaction/components/EmojiPickerPopover";
import { ScheduleSendMenu } from "#/features/schedule/components/ScheduleSendMenu";

import type { FormatKey } from "../utils/format";

import type { Icon } from "@tabler/icons-react";

type MessageInputToolbarProps = {
  isPreview: boolean;
  onTogglePreview: () => void;
  onSubmit: () => void;
  isSendDisabled: boolean;
  isSending: boolean;
  activeFormats: Record<FormatKey, boolean>;
  onFormat: (key: FormatKey) => void;
  onInsertEmoji: (emoji: string) => void;
  // 絵文字を選んだあと、閉じたピッカーからフォーカスを入力欄へ戻す
  onFocusInput: () => void;
  onFileSelect: (files: File[]) => void;
  onShareLocation: () => void;
  onRecord: () => void;
  onCreatePoll: () => void;
  onSchedule: (scheduledAt: Date) => void;
};

const formatButtons: { key: FormatKey; icon: Icon }[] = [
  { icon: IconBold, key: "bold" },
  { icon: IconItalic, key: "italic" },
  { icon: IconStrikethrough, key: "strikethrough" },
  { icon: IconCode, key: "code" },
  { icon: IconLink, key: "link" },
  { icon: IconH1, key: "heading" },
  { icon: IconQuote, key: "quote" },
  { icon: IconList, key: "list" },
  { icon: IconListNumbers, key: "orderedList" },
];

const buttonClassName = "size-7 [&_svg]:size-4";

export const MessageInputToolbar = ({
  isPreview,
  onTogglePreview,
  onSubmit,
  isSendDisabled,
  isSending,
  activeFormats,
  onFormat,
  onInsertEmoji,
  onFocusInput,
  onFileSelect,
  onShareLocation,
  onRecord,
  onCreatePoll,
  onSchedule,
}: MessageInputToolbarProps) => {
  const { t } = useTranslation();
  const navigate = useNavigate();
  // ピッカーは閉じたあとでフォーカスを開いたボタンへ戻すため、そのときに入力欄へ移す
  const hasPickedEmojiRef = useRef(false);

  return (
    <div className="flex items-center gap-px px-1.25 pb-1.25 max-md:gap-1">
      {/* 幅が足りないときは送信まわり以外を横にスクロールさせる */}
      <div className="flex min-w-0 flex-1 items-center gap-px overflow-x-auto [scrollbar-width:none] max-md:gap-1 max-md:mask-r-from-85%">
        {formatButtons.map(({ key, icon: FormatIcon }) => (
          <IconToggleButton
            key={key}
            label={t(`message.composer.${key}`)}
            isDisabled={isPreview}
            isSelected={activeFormats[key]}
            className={buttonClassName}
            onChange={() => {
              onFormat(key);
            }}
          >
            <FormatIcon />
          </IconToggleButton>
        ))}
        <span className="mx-1 h-4 w-px shrink-0 bg-border" />
        <EmojiPickerPopover
          label={t("message.composer.emoji")}
          placement="top start"
          onSelect={(emoji) => {
            hasPickedEmojiRef.current = true;
            onInsertEmoji(emoji);
          }}
          trigger={
            <IconButton
              label={t("message.composer.emoji")}
              isDisabled={isPreview}
              className={buttonClassName}
              onFocus={() => {
                if (hasPickedEmojiRef.current) {
                  hasPickedEmojiRef.current = false;
                  onFocusInput();
                }
              }}
            >
              <IconMoodSmile />
            </IconButton>
          }
        />
        <FileTrigger
          allowsMultiple
          onSelect={(files) => {
            if (files !== null) {
              onFileSelect([...files]);
            }
          }}
        >
          <IconButton label={t("message.composer.attach")} className={buttonClassName}>
            <IconPaperclip />
          </IconButton>
        </FileTrigger>
        <IconButton
          label={t("location.composer.share")}
          className={buttonClassName}
          onPress={onShareLocation}
        >
          <IconMapPin />
        </IconButton>
        <IconButton label={t("recorder.start")} className={buttonClassName} onPress={onRecord}>
          <IconMicrophone />
        </IconButton>
        <IconButton label={t("poll.create")} className={buttonClassName} onPress={onCreatePoll}>
          <IconChartBar />
        </IconButton>
        <IconButton
          label={t("message.composer.help")}
          className={buttonClassName}
          onPress={() => {
            void navigate({ search: openDialog({ dialog: "markdown-help" }), to: "." });
          }}
        >
          <IconHelp />
        </IconButton>
      </div>
      <IconToggleButton
        label={t("message.composer.preview")}
        isSelected={isPreview}
        className={buttonClassName}
        onChange={onTogglePreview}
      >
        <IconEye />
      </IconToggleButton>
      <ScheduleSendMenu isDisabled={isSendDisabled} onSchedule={onSchedule} />
      <IconButton
        label={t("message.composer.send")}
        isDisabled={isSendDisabled}
        onPress={onSubmit}
        className="h-7 w-8 bg-accent text-accent-fg data-disabled:bg-transparent data-hovered:bg-accent-hover data-hovered:text-accent-fg [&_svg]:size-4"
      >
        {isSending ? (
          <IconLoader2 className="animate-spin motion-reduce:animate-none" />
        ) : (
          <IconArrowUp />
        )}
      </IconButton>
    </div>
  );
};
