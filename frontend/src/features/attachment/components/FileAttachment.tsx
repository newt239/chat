import { formatBytes } from "@chat/i18n/format";
import { useMutation } from "@connectrpc/connect-query";
import { IconDownload, IconLoader2 } from "@tabler/icons-react";
import { useTranslation } from "react-i18next";

import { IconButton } from "#/components/ui/IconButton/IconButton";
import { toast } from "#/components/ui/ToastRegion/toast";
import { AttachmentService } from "#/gen/chat/v1/attachment_service_pb";
import { usePreferences } from "#/hooks/usePreferences";
import { openExternal } from "#/lib/platform/openExternal";

import { FileIcon } from "./FileIcon";

import type { MessageAttachment } from "#/gen/chat/v1/message_pb";

type FileAttachmentProps = {
  attachment: MessageAttachment;
};

export const FileAttachment = ({ attachment }: FileAttachmentProps) => {
  const { t } = useTranslation();
  const { locale } = usePreferences();
  const downloadMutation = useMutation(AttachmentService.method.getDownloadUrl);

  const handleDownload = async () => {
    try {
      const { url } = await downloadMutation.mutateAsync({ attachmentId: attachment.id });
      // 署名付き URL は押すたびに発行するため、リンクではなく新しいタブで開く
      await openExternal(url);
    } catch {
      toast(t("attachment.downloadFailed"), { tone: "danger" });
    }
  };

  return (
    <div className="flex w-90 max-w-full items-center gap-2.5 rounded-lg border border-border bg-surface py-2 pr-1.5 pl-2.5 font-sans">
      <FileIcon mimeType={attachment.mimeType} />
      <div className="flex min-w-0 flex-1 flex-col leading-snug">
        <b className="truncate text-body-sm font-semibold">{attachment.fileName}</b>
        <small className="text-caption text-muted">
          {formatBytes(Number(attachment.sizeBytes), locale)}
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
