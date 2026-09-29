import { IconMapPin, IconX } from "@tabler/icons-react";
import { useTranslation } from "react-i18next";

import { IconButton } from "#/components/ui/IconButton";

import { formatCoordinates } from "../utils/externalMapUrl";

import type { MessageLocation } from "#/gen/chat/v1/message_pb";

type PendingLocationProps = {
  location: MessageLocation;
  onRemove: () => void;
};

// 入力欄に添えた、送信前の位置情報
export const PendingLocation = ({ location, onRemove }: PendingLocationProps) => {
  const { t } = useTranslation();
  return (
    <div className="mx-2 mt-2 flex w-fit max-w-[calc(100%-16px)] items-center gap-2 rounded-md border border-border bg-sunken py-1 pr-1 pl-2 font-sans">
      <IconMapPin aria-hidden className="size-4 shrink-0 text-accent-text" />
      <span className="flex min-w-0 flex-col leading-[1.3]">
        <span className="truncate text-xs font-medium">
          {location.label ?? t("location.card.title")}
        </span>
        <span className="truncate font-mono text-[11px] text-muted tabular-nums">
          {formatCoordinates(location)}
        </span>
      </span>
      <IconButton
        label={t("location.composer.remove")}
        onPress={onRemove}
        className="size-6 [&_svg]:size-3.5"
      >
        <IconX />
      </IconButton>
    </div>
  );
};
