import { cn } from "#/components/ui/styles/styles";

import { useCustomEmojiMap } from "../hooks/useCustomEmojis";
import { toCustomEmojiValue } from "../utils/customEmoji";

type CustomEmojiProps = {
  name: string;
  className?: string;
};

// 文字の大きさに合わせて表示する。登録されていない名前は :name: のまま出す
export const CustomEmoji = ({ name, className }: CustomEmojiProps) => {
  const emoji = useCustomEmojiMap().get(name);
  const label = toCustomEmojiValue(name);
  if (emoji === undefined) {
    return <span className={className}>{label}</span>;
  }
  return (
    <img
      src={emoji.imageUrl}
      alt={label}
      title={label}
      loading="lazy"
      draggable={false}
      className={cn("inline-block size-[1.375em] object-contain align-[-0.3em]", className)}
    />
  );
};
