import { Box, Paper } from '@mantine/core';

import { MessageContent } from './MessageContent';

interface MessagePreviewProps {
  content: string;
}

export const MessagePreview = ({ content }: MessagePreviewProps) => {
  return (
    <Paper withBorder p="md" mih={100}>
      {content ? (
        <MessageContent content={content} />
      ) : (
        <Box c="dimmed">プレビューするテキストを入力してください</Box>
      )}
    </Paper>
  );
}
