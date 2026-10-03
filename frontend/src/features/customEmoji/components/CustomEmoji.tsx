import { useCustomEmojiMap } from "../hooks/useCustomEmojis";
import { toCustomEmojiValue } from "../utils/customEmoji";

type CustomEmojiProps = {
  name: string;
};

// 文字の大きさに合わせて表示する。登録されていない名前は :name: のまま出す
export const CustomEmoji = ({ name }: CustomEmojiProps) => {
  const emoji = useCustomEmojiMap().get(name);
  const label = toCustomEmojiValue(name);
  if (emoji === undefined) {
    return <span>{label}</span>;
  }
  return (
    <img
      src={emoji.imageUrl}
      alt={label}
      title={label}
      loading="lazy"
      draggable={false}
      className="inline-block size-[1.375em] object-contain align-[-0.3em]"
    />
  );
};
