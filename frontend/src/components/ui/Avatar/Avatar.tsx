import { oklchToHex } from "@chat/design-tokens/color";

import { useColorMode } from "#/providers/theme/colorMode";

type AvatarProps = {
  name: string;
  src?: string | null;
  size?: number;
  isOnline?: boolean;
};

const hueOf = (seed: string) => {
  let hash = 7;
  for (const char of seed) {
    hash = (hash * 31 + (char.codePointAt(0) ?? 0)) % 3600;
  }
  return hash % 360;
};

const initialOf = (name: string) => {
  const [first] = new Intl.Segmenter().segment(name);
  return first?.segment.toUpperCase();
};

// 画像がなければ名前の頭文字を名前から決めた色相の背景に出す。在席の点の縁は --dot-ring で背景色に合わせる
export const Avatar = ({ name, src, size = 32, isOnline }: AvatarProps) => {
  const isDark = useColorMode() === "dark";
  const hue = hueOf(name);
  const dotSize = Math.max(9, Math.round(size * 0.28));
  return (
    <span
      role="img"
      aria-label={name}
      className="relative inline-grid shrink-0 place-items-center font-sans leading-none font-bold select-none"
      style={{
        backgroundColor: oklchToHex(isDark ? 0.38 : 0.88, isDark ? 0.07 : 0.06, hue),
        borderRadius: Math.round(size * 0.28),
        color: oklchToHex(isDark ? 0.92 : 0.36, 0.09, hue),
        fontSize: Math.round(size * 0.42),
        height: size,
        width: size,
      }}
    >
      {src ? (
        <img src={src} alt="" className="size-full rounded-[inherit] object-cover" />
      ) : (
        <span aria-hidden>{initialOf(name)}</span>
      )}
      {isOnline && (
        <span
          data-online
          className="absolute -right-0.5 -bottom-0.5 rounded-full border-2 border-(--dot-ring,var(--c-surface)) bg-success"
          style={{ height: dotSize, width: dotSize }}
        />
      )}
    </span>
  );
};
