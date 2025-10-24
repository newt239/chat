import { useState } from "react";

import { ActionIcon, Menu, Popover } from "@mantine/core";
import {
  IconBookmark,
  IconBookmarkFilled,
  IconDots,
  IconEdit,
  IconLink,
  IconMessage,
  IconMoodSmile,
  IconTrash,
} from "@tabler/icons-react";

import { EmojiPicker } from "@/features/reaction/components/EmojiPicker";
import { useAddReaction } from "@/features/reaction/hooks/useReactions";
import {
  useAddBookmark,
  useRemoveBookmark,
  useIsBookmarked,
} from "@/features/bookmark/hooks/useBookmarks";

type MessageActionsProps = {
  messageId: string;
  onCopyLink: (messageId: string) => void;
  onCreateThread: (messageId: string) => void;
  onBookmark: (messageId: string) => void;
};

export const MessageActions = ({
  messageId,
  onCopyLink,
  onCreateThread,
  onBookmark,
}: MessageActionsProps) => {
  const [emojiPickerOpened, setEmojiPickerOpened] = useState(false);
  const addReaction = useAddReaction();
  const addBookmark = useAddBookmark();
  const removeBookmark = useRemoveBookmark();
  const isBookmarked = useIsBookmarked(messageId);

  const handleEmojiSelect = async (emoji: string) => {
    await addReaction.mutateAsync({ messageId, emoji });
    setEmojiPickerOpened(false);
  };

  const handleBookmarkToggle = async () => {
    if (isBookmarked) {
      await removeBookmark.mutateAsync({ messageId });
    } else {
      await addBookmark.mutateAsync({ messageId });
    }
    onBookmark(messageId);
  };

  return (
    <div className="absolute right-4 top-2 flex gap-1 rounded-md border bg-white p-1 shadow-sm">
      <Popover
        opened={emojiPickerOpened}
        onChange={setEmojiPickerOpened}
        position="bottom"
        withArrow
      >
        <Popover.Target>
          <ActionIcon
            variant="subtle"
            size="sm"
            onClick={() => setEmojiPickerOpened((o) => !o)}
            title="リアクションを追加"
          >
            <IconMoodSmile size={16} />
          </ActionIcon>
        </Popover.Target>
        <Popover.Dropdown>
          <EmojiPicker onEmojiSelect={handleEmojiSelect} />
        </Popover.Dropdown>
      </Popover>

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
        onClick={handleBookmarkToggle}
        title={isBookmarked ? "ブックマークを削除" : "ブックマークに追加"}
        c={isBookmarked ? "blue" : undefined}
      >
        {isBookmarked ? <IconBookmarkFilled size={16} /> : <IconBookmark size={16} />}
      </ActionIcon>

      <Menu position="bottom-end">
        <Menu.Target>
          <ActionIcon variant="subtle" size="sm">
            <IconDots size={16} />
          </ActionIcon>
        </Menu.Target>
        <Menu.Dropdown>
          <Menu.Item leftSection={<IconLink size={14} />} onClick={() => onCopyLink(messageId)}>
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
