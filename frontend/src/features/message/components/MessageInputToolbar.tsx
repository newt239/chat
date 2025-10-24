import type { RefObject } from 'react';

import { Group, Button, ActionIcon } from '@mantine/core';
import {
  IconBold,
  IconItalic,
  IconStrikethrough,
  IconH1,
  IconLink,
  IconCode,
  IconQuote,
  IconList,
  IconListNumbers,
  IconInfoCircle,
  IconSend,
} from '@tabler/icons-react';

import type { MessageInputMode } from '../hooks/useMessageInputMode';

interface MessageInputToolbarProps {
  mode: MessageInputMode;
  onToggleMode: () => void;
  onSubmit: () => void;
  disabled: boolean;
  loading: boolean;
  textareaRef: RefObject<HTMLTextAreaElement | null>;
}

type FormatPattern = {
  prefix: string;
  suffix: string;
  cursorOffset: number;
};

const FORMAT_PATTERNS: Record<string, FormatPattern> = {
  bold: { prefix: '**', suffix: '**', cursorOffset: 2 },
  italic: { prefix: '_', suffix: '_', cursorOffset: 1 },
  strikethrough: { prefix: '~~', suffix: '~~', cursorOffset: 2 },
  heading: { prefix: '# ', suffix: '', cursorOffset: 2 },
  link: { prefix: '[', suffix: '](url)', cursorOffset: 1 },
  code: { prefix: '`', suffix: '`', cursorOffset: 1 },
  quote: { prefix: '> ', suffix: '', cursorOffset: 2 },
  list: { prefix: '- ', suffix: '', cursorOffset: 2 },
  orderedList: { prefix: '1. ', suffix: '', cursorOffset: 3 },
};

export const MessageInputToolbar = ({
  mode,
  onToggleMode,
  onSubmit,
  disabled,
  loading,
  textareaRef,
}: MessageInputToolbarProps) => {
  const insertFormat = (formatKey: string) => {
    const textarea = textareaRef.current;
    if (!textarea) return;

    const pattern = FORMAT_PATTERNS[formatKey];
    const start = textarea.selectionStart;
    const end = textarea.selectionEnd;
    const selectedText = textarea.value.substring(start, end);
    const beforeText = textarea.value.substring(0, start);
    const afterText = textarea.value.substring(end);

    const newText = beforeText + pattern.prefix + selectedText + pattern.suffix + afterText;
    const newCursorPos = start + pattern.prefix.length + (selectedText.length > 0 ? selectedText.length : 0);

    // テキストエリアの値を更新
    textarea.value = newText;

    // カーソル位置を設定
    textarea.focus();
    textarea.setSelectionRange(newCursorPos, newCursorPos);

    // Reactの変更イベントを手動でトリガー
    const event = new Event('input', { bubbles: true });
    textarea.dispatchEvent(event);
  };

  return (
    <Group gap="xs" mb="xs" justify="space-between">
      <Group gap={4}>
        <ActionIcon
          variant="subtle"
          size="sm"
          color="gray"
          onClick={() => insertFormat('bold')}
          disabled={mode === 'preview'}
        >
          <IconBold size={16} />
        </ActionIcon>
        <ActionIcon
          variant="subtle"
          size="sm"
          color="gray"
          onClick={() => insertFormat('italic')}
          disabled={mode === 'preview'}
        >
          <IconItalic size={16} />
        </ActionIcon>
        <ActionIcon
          variant="subtle"
          size="sm"
          color="gray"
          onClick={() => insertFormat('strikethrough')}
          disabled={mode === 'preview'}
        >
          <IconStrikethrough size={16} />
        </ActionIcon>
        <ActionIcon
          variant="subtle"
          size="sm"
          color="gray"
          onClick={() => insertFormat('heading')}
          disabled={mode === 'preview'}
        >
          <IconH1 size={16} />
        </ActionIcon>
        <ActionIcon
          variant="subtle"
          size="sm"
          color="gray"
          onClick={() => insertFormat('link')}
          disabled={mode === 'preview'}
        >
          <IconLink size={16} />
        </ActionIcon>
        <ActionIcon
          variant="subtle"
          size="sm"
          color="gray"
          onClick={() => insertFormat('code')}
          disabled={mode === 'preview'}
        >
          <IconCode size={16} />
        </ActionIcon>
        <ActionIcon
          variant="subtle"
          size="sm"
          color="gray"
          onClick={() => insertFormat('quote')}
          disabled={mode === 'preview'}
        >
          <IconQuote size={16} />
        </ActionIcon>
        <ActionIcon
          variant="subtle"
          size="sm"
          color="gray"
          onClick={() => insertFormat('list')}
          disabled={mode === 'preview'}
        >
          <IconList size={16} />
        </ActionIcon>
        <ActionIcon
          variant="subtle"
          size="sm"
          color="gray"
          onClick={() => insertFormat('orderedList')}
          disabled={mode === 'preview'}
        >
          <IconListNumbers size={16} />
        </ActionIcon>
        <ActionIcon variant="subtle" size="sm" color="gray" disabled={mode === 'preview'}>
          <IconInfoCircle size={16} />
        </ActionIcon>
      </Group>
      <Group gap="xs">
        <Button
          variant="subtle"
          size="compact-sm"
          onClick={onToggleMode}
          color="gray"
        >
          {mode === 'edit' ? 'プレビュー' : '編集'}
        </Button>
        <ActionIcon
          variant="filled"
          size="lg"
          color="green"
          onClick={onSubmit}
          disabled={disabled}
          loading={loading}
        >
          <IconSend size={18} />
        </ActionIcon>
      </Group>
    </Group>
  );
}
