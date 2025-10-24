import data from "@emoji-mart/data";
import Picker from "@emoji-mart/react";

type EmojiPickerProps = {
  onEmojiSelect: (emoji: string) => void;
}

type EmojiSelectEvent = {
  native: string;
}

export const EmojiPicker = ({ onEmojiSelect }: EmojiPickerProps) => {
  const handleEmojiSelect = (emoji: EmojiSelectEvent) => {
    onEmojiSelect(emoji.native);
  };

  return (
    <Picker
      data={data}
      onEmojiSelect={handleEmojiSelect}
      theme="light"
      locale="ja"
      previewPosition="none"
    />
  );
};
