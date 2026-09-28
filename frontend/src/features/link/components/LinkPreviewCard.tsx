import { IconX } from "@tabler/icons-react";
import { useTranslation } from "react-i18next";

import { IconButton } from "#/components/ui/IconButton";
import { Skeleton } from "#/components/ui/Skeleton";

import { OgpCard } from "./OgpCard";

import type { LinkPreview } from "../types";

type LinkPreviewCardProps = {
  preview: LinkPreview;
  onRemove: () => void;
};

export const LinkPreviewCard = ({ preview, onRemove }: LinkPreviewCardProps) => {
  const { t } = useTranslation();
  const { url, ogpData, isLoading, error } = preview;

  if (isLoading) {
    return (
      <div className="flex w-[min(420px,100%)] flex-col gap-1.5 rounded-[10px] border border-border p-3">
        <Skeleton className="h-3 w-24" />
        <Skeleton className="h-4 w-4/5" />
        <Skeleton className="h-3 w-3/5" />
      </div>
    );
  }

  if (error) {
    return (
      <div className="flex w-[min(420px,100%)] items-center gap-2 rounded-[10px] border border-border py-1.5 pr-1.5 pl-3 text-caption">
        <span className="min-w-0 flex-1 truncate text-danger">
          {t("link.previewFailed")} · <span className="text-muted">{url}</span>
        </span>
        <IconButton label={t("link.remove")} onPress={onRemove} className="size-6 [&_svg]:size-3.5">
          <IconX />
        </IconButton>
      </div>
    );
  }

  return <OgpCard url={url} ogp={ogpData} onRemove={onRemove} />;
};
