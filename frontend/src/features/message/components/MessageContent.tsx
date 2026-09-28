import { Box } from "@mantine/core";

import { LinkPreviewEmbed } from "#/features/link/components/LinkPreviewEmbed";

import { renderMarkdown } from "../utils/markdown/renderer";

import type { Message } from "#/gen/chat/v1/message_pb";

type MessageContentProps = {
  message: Message;
};

export const MessageContent = ({ message }: MessageContentProps) => {
  const { body, links } = message;

  const rendered = renderMarkdown(body);

  return (
    <div className="space-y-2">
      <Box className="message-content prose prose-sm max-w-none">{rendered}</Box>

      {/* リンクプレビューを表示 */}
      {links.length > 0 && (
        <div className="space-y-2">
          {links.map((link) => (
            <LinkPreviewEmbed key={link.id} link={link} />
          ))}
        </div>
      )}
    </div>
  );
};
