import { LinkPreviewEmbed } from "#/features/link/components/LinkPreviewEmbed";

import { markdownClassName } from "../utils/markdown/className";
import { renderMarkdown } from "../utils/markdown/renderer";

import type { Message } from "#/gen/chat/v1/message_pb";

type MessageContentProps = {
  message: Message;
};

export const MessageContent = ({ message }: MessageContentProps) => (
  <>
    <div className={markdownClassName}>{renderMarkdown(message.body)}</div>
    {message.links.map((link) => (
      <LinkPreviewEmbed key={link.id} link={link} />
    ))}
  </>
);
