import { Box, Paper } from "@mantine/core";

import { renderMarkdown } from "../utils/markdown/renderer";

type MessagePreviewProps = {
  content: string;
};

export const MessagePreview = ({ content }: MessagePreviewProps) => {
  return (
    <Paper withBorder p="md" mih={100} className="text-sm">
      {content ? (
        <Box className="message-content prose prose-sm max-w-none">{renderMarkdown(content)}</Box>
      ) : (
        <Box c="dimmed">プレビューするテキストを入力してください</Box>
      )}
    </Paper>
  );
};
