import data from "@emoji-mart/data";
import en from "@emoji-mart/data/i18n/en.json";
import ja from "@emoji-mart/data/i18n/ja.json";
import Picker from "@emoji-mart/react";
import { useTranslation } from "react-i18next";

import { useCustomEmojiMap } from "#/features/customEmoji/hooks/useCustomEmojis";
import { toCustomEmojiValue } from "#/features/customEmoji/utils/customEmoji";
import { usePreferences } from "#/hooks/usePreferences";
import { useColorMode } from "#/providers/theme/colorMode";

type EmojiPickerProps = {
  onEmojiSelect: (emoji: string) => void;
};

// カスタム絵文字は native を持たず、name に登録名が入る
type EmojiSelectEvent = {
  native?: string;
  name: string;
};

// 翻訳を同期的に渡す。取得を待つ間に初期化が重なると、カスタムの分類が二重に追加される
const i18n = { en, ja };

// 標準の絵文字と ID が重ならないよう接頭辞を付ける
const CUSTOM_ID_PREFIX = "custom-";

export const EmojiPicker = ({ onEmojiSelect }: EmojiPickerProps) => {
  const { t } = useTranslation();
  const { locale } = usePreferences();
  const colorMode = useColorMode();
  const customEmojis = useCustomEmojiMap();
  const custom = [
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
  ];

  return (
    <Picker
      data={data}
      custom={custom}
      onEmojiSelect={(emoji: EmojiSelectEvent) => {
        onEmojiSelect(emoji.native ?? toCustomEmojiValue(emoji.name));
      }}
      theme={colorMode}
      locale={locale}
      i18n={i18n[locale]}
      previewPosition="none"
    />
  );
};
