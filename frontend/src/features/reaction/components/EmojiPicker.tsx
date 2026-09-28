import data from "@emoji-mart/data";
import Picker from "@emoji-mart/react";
import { useAtomValue } from "jotai";

import { preferencesAtom } from "#/providers/store/preferences";
import { useColorMode } from "#/providers/theme/colorMode";

type EmojiPickerProps = {
  onEmojiSelect: (emoji: string) => void;
};

type EmojiSelectEvent = {
  native: string;
};

export const EmojiPicker = ({ onEmojiSelect }: EmojiPickerProps) => {
  const { locale } = useAtomValue(preferencesAtom);
  const colorMode = useColorMode();

  return (
    <Picker
      data={data}
      onEmojiSelect={(emoji: EmojiSelectEvent) => {
        onEmojiSelect(emoji.native);
      }}
      theme={colorMode}
      locale={locale}
      previewPosition="none"
    />
  );
};
