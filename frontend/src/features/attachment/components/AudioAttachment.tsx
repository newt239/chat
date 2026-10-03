import { useAttachmentUrl } from "../hooks/useAttachmentUrl";

import type { MessageAttachment } from "#/gen/chat/v1/message_pb";

type AudioAttachmentProps = {
  attachment: MessageAttachment;
};

export const AudioAttachment = ({ attachment }: AudioAttachmentProps) => {
  const { data: url, isStale, refetch } = useAttachmentUrl(attachment.id, false);

  return (
    <div className="flex w-100 max-w-full flex-col gap-1.5 rounded-xl border border-border bg-surface p-2 font-sans">
      <b className="truncate px-1 text-label font-semibold text-text">{attachment.fileName}</b>
      {/* oxlint-disable-next-line jsx-a11y/media-has-caption -- 利用者が上げた音声に字幕はない */}
      <audio
        controls
        preload="none"
        src={url}
        aria-label={attachment.fileName}
        // 署名付き URL の期限が切れて読めなかったら取り直す
        onError={() => {
          if (isStale) {
            void refetch();
          }
        }}
        className="w-full"
      />
    </div>
  );
};
