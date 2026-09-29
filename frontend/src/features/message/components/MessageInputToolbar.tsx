import {
  IconArrowUp,
  IconBold,
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
  IconPaperclip,
  IconQuote,
  IconStrikethrough,
} from "@tabler/icons-react";
import { useNavigate } from "@tanstack/react-router";
import { FileTrigger } from "react-aria-components";
import { useTranslation } from "react-i18next";

import { IconButton } from "#/components/ui/IconButton";
import { cn } from "#/components/ui/styles";
import { openDialog } from "#/features/layout/utils/overlaySearch";
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
  onFileSelect: (files: File[]) => void;
  onShareLocation: () => void;
  onRecord: () => void;
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
const activeClassName = "bg-accent-soft text-accent-text data-hovered:bg-accent-soft";

export const MessageInputToolbar = ({
  isPreview,
  onTogglePreview,
  onSubmit,
  isSendDisabled,
  isSending,
  activeFormats,
  onFormat,
  onFileSelect,
  onShareLocation,
  onRecord,
  onSchedule,
}: MessageInputToolbarProps) => {
  const { t } = useTranslation();
  const navigate = useNavigate();

  return (
    <div className="@container flex items-center gap-px px-[5px] pb-[5px]">
      {/* モバイルやスレッドの欄は幅が足りないため、書式のボタンを出さない */}
      <div className="hidden items-center gap-px @lg:flex">
        {formatButtons.map(({ key, icon: FormatIcon }) => (
          <IconButton
            key={key}
            label={t(`message.composer.${key}`)}
            isDisabled={isPreview}
            className={cn(buttonClassName, activeFormats[key] && activeClassName)}
            onPress={() => {
              onFormat(key);
            }}
          >
            <FormatIcon />
          </IconButton>
        ))}
        <span className="mx-1 h-4 w-px bg-border" />
      </div>
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
      <IconButton
        label={t("message.composer.help")}
        className={cn(buttonClassName, "hidden @lg:inline-grid")}
        onPress={() => {
          void navigate({ search: openDialog({ dialog: "markdown-help" }), to: "." });
        }}
      >
        <IconHelp />
      </IconButton>
      <span className="flex-1" />
      <IconButton
        label={t("message.composer.preview")}
        className={cn(buttonClassName, isPreview && activeClassName)}
        onPress={onTogglePreview}
      >
        <IconEye />
      </IconButton>
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
