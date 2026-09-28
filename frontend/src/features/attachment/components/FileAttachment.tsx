import { IconDownload, IconLoader2 } from "@tabler/icons-react";
import { useTranslation } from "react-i18next";

import { IconButton } from "#/components/ui/IconButton";
import { toast } from "#/components/ui/toast";

import { useDownloadUrl } from "../api/client";
import { formatFileSize } from "../utils/validator";
import { FileIcon } from "./FileIcon";

import type { MessageAttachment } from "#/gen/chat/v1/message_pb";

type FileAttachmentProps = {
  attachment: MessageAttachment;
};

export const FileAttachment = ({ attachment }: FileAttachmentProps) => {
  const { t } = useTranslation();
  const downloadMutation = useDownloadUrl();

  const handleDownload = async () => {
    try {
      const { url } = await downloadMutation.mutateAsync({ attachmentId: attachment.id });
      // 署名付き URL は押すたびに発行するため、リンクではなく新しいタブで開く
      window.open(url, "_blank", "noopener,noreferrer");
    } catch {
      toast(t("attachment.downloadFailed"), { tone: "danger" });
    }
  };

  return (
    <div className="flex w-[min(360px,100%)] items-center gap-2.5 rounded-[10px] border border-border bg-surface py-2 pr-1.5 pl-2.5 font-sans">
      <FileIcon mimeType={attachment.mimeType} />
      <div className="flex min-w-0 flex-1 flex-col leading-[1.35]">
        <b className="truncate text-[13px] font-semibold">{attachment.fileName}</b>
        <small className="text-[11.5px] text-muted">
          {formatFileSize(Number(attachment.sizeBytes))}
        </small>
      </div>
      <IconButton
        label={t("attachment.download")}
        isDisabled={downloadMutation.isPending}
        onPress={() => {
          void handleDownload();
        }}
      >
        {downloadMutation.isPending ? (
          <IconLoader2 className="animate-spin motion-reduce:animate-none" />
        ) : (
          <IconDownload />
        )}
      </IconButton>
    </div>
  );
};
