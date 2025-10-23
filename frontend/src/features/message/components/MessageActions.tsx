import { ActionIcon, Menu } from "@mantine/core";
import {
  IconBookmark,
  IconDots,
  IconEdit,
  IconLink,
  IconMessage,
  IconMoodSmile,
  IconTrash,
} from "@tabler/icons-react";

interface MessageActionsProps {
  messageId: string;
  onCopyLink: (messageId: string) => void;
  onAddReaction: (messageId: string) => void;
  onCreateThread: (messageId: string) => void;
  onBookmark: (messageId: string) => void;
}

export const MessageActions = ({
  messageId,
  onCopyLink,
  onAddReaction,
  onCreateThread,
  onBookmark,
}: MessageActionsProps) => {
  return (
    <div className="absolute right-4 top-2 flex gap-1 rounded-md border bg-white p-1 shadow-sm">
      <ActionIcon
        variant="subtle"
        size="sm"
        onClick={() => onAddReaction(messageId)}
        title="リアクションを追加"
      >
        <IconMoodSmile size={16} />
      </ActionIcon>

      <ActionIcon
        variant="subtle"
        size="sm"
        onClick={() => onCreateThread(messageId)}
        title="スレッドで返信"
      >
        <IconMessage size={16} />
      </ActionIcon>

      <ActionIcon
        variant="subtle"
        size="sm"
        onClick={() => onBookmark(messageId)}
        title="ブックマークに追加"
      >
        <IconBookmark size={16} />
      </ActionIcon>

      <Menu position="bottom-end">
        <Menu.Target>
          <ActionIcon variant="subtle" size="sm">
            <IconDots size={16} />
          </ActionIcon>
        </Menu.Target>
        <Menu.Dropdown>
          <Menu.Item
            leftSection={<IconLink size={14} />}
            onClick={() => onCopyLink(messageId)}
          >
            リンクをコピー
          </Menu.Item>
          <Menu.Item leftSection={<IconEdit size={14} />}>メッセージを編集</Menu.Item>
          <Menu.Item leftSection={<IconTrash size={14} />} c="red">
            メッセージを削除
          </Menu.Item>
        </Menu.Dropdown>
      </Menu>
    </div>
  );
};
