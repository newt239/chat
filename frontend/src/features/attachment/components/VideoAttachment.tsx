import { useAttachmentUrl } from "../hooks/useAttachmentUrl";

import type { MessageAttachment } from "#/gen/chat/v1/message_pb";

type VideoAttachmentProps = {
  attachment: MessageAttachment;
};

const MAX_WIDTH = 420;
const MAX_HEIGHT = 320;

export const VideoAttachment = ({ attachment }: VideoAttachmentProps) => {
  const { width, height, thumbnail } = attachment.media ?? {};
  const ratio = width && height ? width / height : 16 / 9;
  const { data: url, isStale, refetch } = useAttachmentUrl(attachment.id, false);
  const { data: posterUrl } = useAttachmentUrl(thumbnail ? attachment.id : null, true);

  return (
    // oxlint-disable-next-line jsx-a11y/media-has-caption -- 利用者が上げた動画に字幕はない
    <video
      controls
      preload="none"
      src={url}
      poster={posterUrl}
      aria-label={attachment.fileName}
      // 署名付き URL の期限が切れて読めなかったら取り直す
      onError={() => {
        if (isStale) {
          void refetch();
        }
      }}
      className="max-w-full rounded-lg border border-border bg-media"
      style={{ aspectRatio: ratio, width: Math.min(MAX_WIDTH, Math.round(MAX_HEIGHT * ratio)) }}
    />
  );
};
