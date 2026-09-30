import { useMemo } from "react";

import data from "@emoji-mart/data";
import Picker from "@emoji-mart/react";
import { useAtomValue } from "jotai";
import { useTranslation } from "react-i18next";

import { useCustomEmojiMap } from "#/features/customEmoji/hooks/useCustomEmojis";
import { toCustomEmojiValue } from "#/features/customEmoji/utils/customEmoji";
import { preferencesAtom } from "#/providers/store/preferences";
import { useColorMode } from "#/providers/theme/colorMode";

type EmojiPickerProps = {
  onEmojiSelect: (emoji: string) => void;
};

// カスタム絵文字は native を持たず、name に登録名が入る
type EmojiSelectEvent = {
  native?: string;
  name: string;
};

// 標準の絵文字と ID が重ならないよう接頭辞を付ける
const CUSTOM_ID_PREFIX = "custom-";

export const EmojiPicker = ({ onEmojiSelect }: EmojiPickerProps) => {
  const { t } = useTranslation();
  const { locale } = useAtomValue(preferencesAtom);
  const colorMode = useColorMode();
  const customEmojis = useCustomEmojiMap();
  const custom = useMemo(
    () => [
      {
        emojis: [...customEmojis.values()].map((emoji) => ({
          id: `${CUSTOM_ID_PREFIX}${emoji.name}`,
          keywords: [emoji.name],
          name: emoji.name,
          skins: [{ src: emoji.imageUrl }],
        })),
        id: "custom",
        name: t("reaction.picker.custom"),
      },
    ],
    [customEmojis, t],
  );

  return (
    <Picker
      data={data}
      custom={custom}
      onEmojiSelect={(emoji: EmojiSelectEvent) => {
        onEmojiSelect(emoji.native ?? toCustomEmojiValue(emoji.name));
      }}
      theme={colorMode}
      locale={locale}
      previewPosition="none"
    />
  );
};
