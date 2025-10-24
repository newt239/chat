import { Group, Button } from '@mantine/core';
import { IconEdit, IconEye } from '@tabler/icons-react';

import type { MessageInputMode } from '../hooks/useMessageInputMode';

interface MessageInputToolbarProps {
  mode: MessageInputMode;
  onToggleMode: () => void;
}

export const MessageInputToolbar = ({
  mode,
  onToggleMode,
}: MessageInputToolbarProps) => {
  return (
    <Group gap="xs" mb="xs">
      <Button
        variant={mode === 'edit' ? 'filled' : 'light'}
        size="xs"
        leftSection={<IconEdit size={14} />}
        onClick={mode === 'preview' ? onToggleMode : undefined}
      >
        編集
      </Button>
      <Button
        variant={mode === 'preview' ? 'filled' : 'light'}
        size="xs"
        leftSection={<IconEye size={14} />}
        onClick={mode === 'edit' ? onToggleMode : undefined}
      >
        プレビュー
      </Button>
    </Group>
  );
}
