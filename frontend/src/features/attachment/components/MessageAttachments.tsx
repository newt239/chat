import { mediaKindOf } from "../utils/mediaKind";
import { AudioAttachment } from "./AudioAttachment";
import { FileAttachment } from "./FileAttachment";
import { ImageGallery } from "./ImageGallery";
import { VideoAttachment } from "./VideoAttachment";

import type { Message } from "#/gen/chat/v1/message_pb";

type MessageAttachmentsProps = {
  message: Message;
};

// 画像はまとめてグリッドに、動画・音声はブラウザのプレイヤー、それ以外はファイルのカードで並べる
export const MessageAttachments = ({ message }: MessageAttachmentsProps) => {
  const attachments = message.attachments.map((attachment) => ({
    attachment,
    kind: mediaKindOf(attachment.mimeType),
  }));
  const images = attachments
    .filter(({ kind }) => kind === "image")
    .map(({ attachment }) => attachment);

  return (
    <>
      {images.length > 0 && <ImageGallery images={images} message={message} />}
      {attachments.map(({ attachment, kind }) => {
        if (kind === "video") {
          return <VideoAttachment key={attachment.id} attachment={attachment} />;
        }
        if (kind === "audio") {
          return <AudioAttachment key={attachment.id} attachment={attachment} />;
        }
        return kind === "file" ? (
          <FileAttachment key={attachment.id} attachment={attachment} />
        ) : null;
      })}
    </>
  );
};
