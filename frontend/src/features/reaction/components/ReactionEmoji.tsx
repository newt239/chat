import { CustomEmoji } from "#/features/customEmoji/components/CustomEmoji";
import { customEmojiName } from "#/features/customEmoji/utils/customEmoji";

type ReactionEmojiProps = {
  emoji: string;
};

// リアクションの値。:name: はカスタム絵文字の画像にする
export const ReactionEmoji = ({ emoji }: ReactionEmojiProps) => {
  const name = customEmojiName(emoji);
  return name === null ? emoji : <CustomEmoji name={name} />;
};
