import { useState } from "react";

import EmojiPicker from "@emoji-mart/react";
import { ActionIcon, Popover } from "@mantine/core";
import { IconPlus } from "@tabler/icons-react";

type AddAnotherEmojiButtonProps = {
  onClick?: (emoji: string) => void;
};

const AddAnotherEmojiButton = ({ onClick }: AddAnotherEmojiButtonProps) => {
  const [emojiPickerOpened, setEmojiPickerOpened] = useState(false);

  const handleEmojiSelect = (emoji: string) => {
    onClick?.(emoji);
    setEmojiPickerOpened(false);
  };

  return (
    <Popover opened={emojiPickerOpened} onChange={setEmojiPickerOpened}>
      <Popover.Target>
        <ActionIcon
          onClick={() => setEmojiPickerOpened((o) => !o)}
          variant="outline"
          color="cyan"
          radius="full"
        >
          <IconPlus size={16} />
        </ActionIcon>
      </Popover.Target>
      <Popover.Dropdown>
        <EmojiPicker onEmojiSelect={handleEmojiSelect} />
      </Popover.Dropdown>
    </Popover>
  );
};

export default AddAnotherEmojiButton;
