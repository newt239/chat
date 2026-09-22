import type { RefObject } from "react";
import { useRef, useState, useEffect } from "react";

import { Group, ActionIcon, Menu } from "@mantine/core";
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
  IconPaperclip,
  IconDots,
  IconEye,
} from "@tabler/icons-react";

import type { MessageInputMode } from "../hooks/useMessageInputMode";

type MessageInputToolbarProps = {
  mode: MessageInputMode;
  onToggleMode: () => void;
  onSubmit: () => void;
  disabled: boolean;
  loading: boolean;
  textareaRef: RefObject<HTMLTextAreaElement | null>;
  onFileSelect: (files: File[]) => void;
};

type FormatPattern = {
  prefix: string;
  suffix: string;
  cursorOffset: number;
};

const FORMAT_PATTERNS: Record<string, FormatPattern> = {
  bold: { cursorOffset: 2, prefix: "**", suffix: "**" },
  code: { cursorOffset: 1, prefix: "`", suffix: "`" },
  heading: { cursorOffset: 2, prefix: "# ", suffix: "" },
  italic: { cursorOffset: 1, prefix: "_", suffix: "_" },
  link: { cursorOffset: 1, prefix: "[", suffix: "](url)" },
  list: { cursorOffset: 2, prefix: "- ", suffix: "" },
  orderedList: { cursorOffset: 3, prefix: "1. ", suffix: "" },
  quote: { cursorOffset: 2, prefix: "> ", suffix: "" },
  strikethrough: { cursorOffset: 2, prefix: "~~", suffix: "~~" },
};

type ActiveFormats = {
  bold: boolean;
  italic: boolean;
  strikethrough: boolean;
  heading: boolean;
  link: boolean;
  code: boolean;
  quote: boolean;
  list: boolean;
  orderedList: boolean;
};

export const MessageInputToolbar = ({
  mode,
  onToggleMode,
  onSubmit,
  disabled,
  loading,
  textareaRef,
  onFileSelect,
}: MessageInputToolbarProps) => {
  const fileInputRef = useRef<HTMLInputElement>(null);
  const [activeFormats, setActiveFormats] = useState<ActiveFormats>({
    bold: false,
    code: false,
    heading: false,
    italic: false,
    link: false,
    list: false,
    orderedList: false,
    quote: false,
    strikethrough: false,
  });

  // カーソル位置の変更を監視してアクティブなフォーマットを更新
  useEffect(() => {
    const textarea = textareaRef.current;
    if (!textarea) {
      return undefined;
    }

    const updateActiveFormats = () => {
      const start = textarea.selectionStart;
      const end = textarea.selectionEnd;
      const text = textarea.value;

      // カーソル位置の前後のテキストを取得
      const beforeCursor = text.slice(0, start);
      const afterCursor = text.slice(end);

      // 現在の行を取得
      const lineStart = beforeCursor.lastIndexOf("\n") + 1;
      const lineEnd = text.indexOf("\n", end);
      const currentLine = text.slice(lineStart, lineEnd === -1 ? text.length : lineEnd);

      // 各フォーマットがアクティブかチェック
      const newActiveFormats: ActiveFormats = {
        bold: /\*\*[^*]*$/.test(beforeCursor) && /^[^*]*\*\*/.test(afterCursor),
        code: /`[^`]*$/.test(beforeCursor) && /^[^`]*`/.test(afterCursor),
        heading: currentLine.trimStart().startsWith("#"),
        italic: /_[^_]*$/.test(beforeCursor) && /^[^_]*_/.test(afterCursor),
        link: /\[[^\]]*$/.test(beforeCursor) && /^[^\]]*\]/.test(afterCursor),
        list: /^-\s/.test(currentLine.trimStart()),
        orderedList: /^\d+\.\s/.test(currentLine.trimStart()),
        quote: currentLine.trimStart().startsWith(">"),
        strikethrough: /~~[^~]*$/.test(beforeCursor) && /^[^~]*~~/.test(afterCursor),
      };

      setActiveFormats(newActiveFormats);
    };

    textarea.addEventListener("selectionchange", updateActiveFormats);
    textarea.addEventListener("input", updateActiveFormats);
    textarea.addEventListener("click", updateActiveFormats);
    textarea.addEventListener("keyup", updateActiveFormats);

    return () => {
      textarea.removeEventListener("selectionchange", updateActiveFormats);
      textarea.removeEventListener("input", updateActiveFormats);
      textarea.removeEventListener("click", updateActiveFormats);
      textarea.removeEventListener("keyup", updateActiveFormats);
    };
  }, [textareaRef]);

  const insertFormat = (formatKey: string) => {
    const textarea = textareaRef.current;
    if (!textarea) {
      return;
    }

    const pattern = FORMAT_PATTERNS[formatKey];
    if (!pattern) {
      return;
    }

    const start = textarea.selectionStart;
    const end = textarea.selectionEnd;
    const selectedText = textarea.value.slice(start, end);
    const beforeText = textarea.value.slice(0, start);
    const afterText = textarea.value.slice(end);

    const newText = beforeText + pattern.prefix + selectedText + pattern.suffix + afterText;
    const newCursorPos =
      start + pattern.prefix.length + (selectedText.length > 0 ? selectedText.length : 0);

    // テキストエリアの値を更新
    textarea.value = newText;

    // カーソル位置を設定
    textarea.focus();
    textarea.setSelectionRange(newCursorPos, newCursorPos);

    // Reactの変更イベントを手動でトリガー
    const event = new Event("input", { bubbles: true });
    textarea.dispatchEvent(event);
  };

  const handleFileClick = () => {
    fileInputRef.current?.click();
  };

  const handleFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const files = [...(e.target.files ?? [])];
    if (files.length > 0) {
      onFileSelect(files);
    }
    // 同じファイルを再選択できるようにリセット
    e.target.value = "";
  };

  return (
    <Group gap="xs" mt="xs" justify="space-between">
      <Group gap={4}>
        <ActionIcon
          variant="subtle"
          size="lg"
          color="gray"
          title="ファイルを添付"
          onClick={handleFileClick}
          disabled={mode === "preview"}
        >
          <IconPaperclip size={16} />
        </ActionIcon>
        <input
          ref={fileInputRef}
          type="file"
          multiple
          onChange={handleFileChange}
          className="hidden"
          disabled={mode === "preview"}
        />

        <ActionIcon
          variant={activeFormats.bold ? "filled" : "subtle"}
          size="lg"
          color={activeFormats.bold ? "blue" : "gray"}
          onClick={() => {
            insertFormat("bold");
          }}
          disabled={mode === "preview"}
        >
          <IconBold size={16} />
        </ActionIcon>

        <ActionIcon
          variant={activeFormats.italic ? "filled" : "subtle"}
          size="lg"
          color={activeFormats.italic ? "blue" : "gray"}
          onClick={() => {
            insertFormat("italic");
          }}
          disabled={mode === "preview"}
        >
          <IconItalic size={16} />
        </ActionIcon>

        <ActionIcon
          variant={activeFormats.strikethrough ? "filled" : "subtle"}
          size="lg"
          color={activeFormats.strikethrough ? "blue" : "gray"}
          onClick={() => {
            insertFormat("strikethrough");
          }}
          disabled={mode === "preview"}
        >
          <IconStrikethrough size={16} />
        </ActionIcon>

        <ActionIcon
          variant={activeFormats.heading ? "filled" : "subtle"}
          size="lg"
          color={activeFormats.heading ? "blue" : "gray"}
          onClick={() => {
            insertFormat("heading");
          }}
          disabled={mode === "preview"}
        >
          <IconH1 size={16} />
        </ActionIcon>

        <ActionIcon
          variant={activeFormats.link ? "filled" : "subtle"}
          size="lg"
          color={activeFormats.link ? "blue" : "gray"}
          onClick={() => {
            insertFormat("link");
          }}
          disabled={mode === "preview"}
        >
          <IconLink size={16} title="リンクを挿入" />
        </ActionIcon>

        <ActionIcon
          variant={activeFormats.code ? "filled" : "subtle"}
          size="lg"
          color={activeFormats.code ? "blue" : "gray"}
          onClick={() => {
            insertFormat("code");
          }}
          disabled={mode === "preview"}
        >
          <IconCode size={16} title="コードを挿入" />
        </ActionIcon>

        <ActionIcon
          variant={activeFormats.quote ? "filled" : "subtle"}
          size="lg"
          color={activeFormats.quote ? "blue" : "gray"}
          onClick={() => {
            insertFormat("quote");
          }}
          disabled={mode === "preview"}
        >
          <IconQuote size={16} title="引用を挿入" />
        </ActionIcon>

        <ActionIcon
          variant={activeFormats.list ? "filled" : "subtle"}
          size="lg"
          color={activeFormats.list ? "blue" : "gray"}
          onClick={() => {
            insertFormat("list");
          }}
          disabled={mode === "preview"}
        >
          <IconList size={16} title="箇条書きを挿入" />
        </ActionIcon>

        <ActionIcon
          variant={activeFormats.orderedList ? "filled" : "subtle"}
          size="lg"
          color={activeFormats.orderedList ? "blue" : "gray"}
          onClick={() => {
            insertFormat("orderedList");
          }}
          disabled={mode === "preview"}
        >
          <IconListNumbers size={16} title="番号付きリストを挿入" />
        </ActionIcon>
        <Menu shadow="md" width={200}>
          <Menu.Target>
            <ActionIcon variant="subtle" size="lg" color="gray" disabled={mode === "preview"}>
              <IconDots size={16} title="その他のオプション" />
            </ActionIcon>
          </Menu.Target>

          <Menu.Dropdown>
            <Menu.Label>その他</Menu.Label>
            <Menu.Item leftSection={<IconInfoCircle size={16} />}>ヘルプ</Menu.Item>
          </Menu.Dropdown>
        </Menu>
      </Group>
      <Group gap="xs">
        <ActionIcon
          variant={mode === "preview" ? "filled" : "subtle"}
          size="lg"
          onClick={onToggleMode}
          title="プレビューモードに切り替え"
        >
          <IconEye size={16} />
        </ActionIcon>
        <ActionIcon
          variant="filled"
          size="lg"
          color="green"
          onClick={onSubmit}
          disabled={disabled}
          loading={loading}
          title="メッセージを送信"
        >
          <IconSend size={18} />
        </ActionIcon>
      </Group>
    </Group>
  );
};
