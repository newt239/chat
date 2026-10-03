import { CUSTOM_EMOJI_IN_TEXT } from "#/features/customEmoji/utils/customEmoji";

const MAX_EMOJI = 8;
// 絵文字・肌の色・国旗・異体字セレクタ・ZWJ・キーキャップと空白だけの文字列
const EMOJI_ONLY =
  /^(?:\p{Extended_Pictographic}|\p{Emoji_Modifier}|\p{Regional_Indicator}|\u{FE0F}|\u{200D}|\u{20E3}|\s)+$/u;

const segmenter = new Intl.Segmenter(undefined, { granularity: "grapheme" });

// 絵文字だけの短い投稿は大きく表示する。登録済みのカスタム絵文字も 1 文字の絵文字として数える
export const isJumboEmoji = (body: string, customEmojiNames: ReadonlyMap<string, object>) => {
  const text = body
    .replaceAll(CUSTOM_EMOJI_IN_TEXT, (token, name: string) =>
      customEmojiNames.has(name) ? "😀" : token,
    )
    .replaceAll(/\s/g, "");
  return text !== "" && EMOJI_ONLY.test(text) && [...segmenter.segment(text)].length <= MAX_EMOJI;
};
