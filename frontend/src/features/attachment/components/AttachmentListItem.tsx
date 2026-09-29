import { IconX } from "@tabler/icons-react";
import { Button } from "react-aria-components";
import { useTranslation } from "react-i18next";

import { focusRing } from "#/components/ui/styles/styles";

import { formatFileSize } from "../utils/validator";
import { FileIcon } from "./FileIcon";

import type { PendingAttachment } from "../api/types";

type AttachmentListItemProps = {
  attachment: PendingAttachment;
  onRemove: () => void;
};

export const AttachmentListItem = ({ attachment, onRemove }: AttachmentListItemProps) => {
  const { t } = useTranslation();
  const { file, state } = attachment;

  return (
    <div className="relative flex w-[200px] items-center gap-2 rounded-md border border-border bg-sunken p-1.5 font-sans">
      <FileIcon mimeType={file.type} />
      <div className="flex min-w-0 flex-1 flex-col gap-0.5 leading-[1.3]">
        <span className="truncate text-xs font-medium">{file.name}</span>
        <span className="truncate text-[11px] text-muted">
          {formatFileSize(file.size)}
          {state.status === "uploading" && ` · ${state.progress}%`}
          {state.status === "completed" && ` · ${t("attachment.completed")}`}
          {state.status === "error" && (
            <span className="text-danger"> · {t("attachment.failed", { error: state.error })}</span>
          )}
        </span>
        {state.status === "uploading" && (
          <span className="h-1 overflow-hidden rounded-full bg-border">
            <span
              className="block h-full rounded-full bg-accent transition-[width] motion-reduce:transition-none"
              style={{ width: `${state.progress}%` }}
            />
          </span>
        )}
      </div>
      <Button
        aria-label={t("attachment.remove")}
        onPress={onRemove}
        isDisabled={state.status === "uploading" || state.status === "presigning"}
        className={`absolute -top-1.5 -right-1.5 grid size-[18px] place-items-center rounded-full bg-text text-surface data-disabled:opacity-40 [&_svg]:size-[11px] ${focusRing}`}
      >
        <IconX aria-hidden stroke={2.5} />
      </Button>
    </div>
  );
};
