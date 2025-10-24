import { Box } from "@mantine/core";

import { LinkPreviewEmbed } from "../../link/components/LinkPreviewEmbed";
import { renderMarkdown } from "../utils/markdown/renderer";

import type { MessageWithUser } from "../types";

interface MessageContentProps {
  message: MessageWithUser;
}

export const MessageContent = ({ message }: MessageContentProps) => {
  const { body, mentions, groups, links } = message;

  // メンションをハイライトするための処理
  const processMentions = (text: string) => {
    let processedText = text;

    // ユーザーメンションをハイライト
    mentions.forEach((mention) => {
      const mentionPattern = new RegExp(`@${mention.displayName}`, "gi");
      processedText = processedText.replace(
        mentionPattern,
        `<span class="mention-user bg-blue-100 text-blue-800 px-1 rounded">@${mention.displayName}</span>`
      );
    });

    // グループメンションをハイライト
    groups.forEach((group) => {
      const mentionPattern = new RegExp(`@${group.name}`, "gi");
      processedText = processedText.replace(
        mentionPattern,
        `<span class="mention-group bg-green-100 text-green-800 px-1 rounded">@${group.name}</span>`
      );
    });

    return processedText;
  };

  const processedContent = processMentions(body);
  const rendered = renderMarkdown(processedContent);

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
