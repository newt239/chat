import { Box } from '@mantine/core';

import { renderMarkdown } from '../utils/markdown/renderer';

interface MessageContentProps {
  content: string;
}

export const MessageContent = ({ content }: MessageContentProps) => {
  const rendered = renderMarkdown(content);

  return (
    <Box className="message-content prose prose-sm max-w-none">
      {rendered}
    </Box>
  );
}
