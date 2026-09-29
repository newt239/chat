import { IconExternalLink, IconMapPin } from "@tabler/icons-react";
import { Link } from "react-aria-components";
import { useTranslation } from "react-i18next";

import { focusRing } from "#/components/ui/styles";

import { externalMapUrl, formatCoordinates } from "../utils/externalMapUrl";
import { LocationMap } from "./LocationMap";

import type { MessageLocation } from "#/gen/chat/v1/message_pb";

type MessageLocationCardProps = {
  location: MessageLocation;
};

export const MessageLocationCard = ({ location }: MessageLocationCardProps) => {
  const { t } = useTranslation();
  return (
    <div className="flex w-[min(400px,100%)] flex-col overflow-hidden rounded-xl border border-border bg-surface font-sans">
      <LocationMap location={location} className="h-[180px] w-full" />
      <div className="flex items-center gap-2 px-3 py-2">
        <IconMapPin aria-hidden className="size-4 shrink-0 text-accent-text" />
        <span className="flex min-w-0 flex-1 flex-col leading-[1.35]">
          <b className="truncate text-[13px] font-semibold text-text">
            {location.label ?? t("location.card.title")}
          </b>
          <small className="truncate font-mono text-[11px] text-subtle tabular-nums">
            {formatCoordinates(location)}
            {location.accuracyMeters !== undefined &&
              ` · ${t("location.accuracy", { meters: Math.round(location.accuracyMeters) })}`}
          </small>
        </span>
        <Link
          href={externalMapUrl(location)}
          target="_blank"
          rel="noopener noreferrer"
          className={`inline-flex shrink-0 items-center gap-1 rounded-sm text-caption text-accent-text no-underline data-hovered:underline [&_svg]:size-3.5 ${focusRing}`}
        >
          {t("location.card.openExternal")}
          <IconExternalLink aria-hidden />
        </Link>
      </div>
    </div>
  );
};
