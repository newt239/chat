import { cn } from "#/components/ui/styles/styles";
import { LinkPreviewEmbed } from "#/features/link/components/LinkPreviewEmbed";

import { isJumboEmoji } from "../utils/isJumboEmoji";
import { jumboClassName, markdownClassName } from "../utils/markdown/className";
import { renderMarkdown } from "../utils/markdown/renderer";

import type { Message } from "#/gen/chat/v1/message_pb";

type MessageContentProps = {
  message: Message;
};

export const MessageContent = ({ message }: MessageContentProps) => (
  <>
    {message.body !== "" && (
      <div className={cn(markdownClassName, isJumboEmoji(message.body) && jumboClassName)}>
        {renderMarkdown(message.body)}
      </div>
    )}
    {message.links.map((link) => (
      <LinkPreviewEmbed key={link.id} link={link} />
    ))}
  </>
);
