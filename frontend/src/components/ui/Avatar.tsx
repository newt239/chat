import { oklchToHex } from "@chat/design-tokens";

import { useColorMode } from "#/providers/theme/colorMode";

import { cn } from "./styles";

type Presence = "online" | "away" | "offline";

type AvatarProps = {
  name: string;
  src?: string | null;
  size?: number;
  presence?: Presence;
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

const presenceClassNames: Record<Presence, string> = {
  away: "bg-(--dot-ring,var(--c-surface)) shadow-[inset_0_0_0_1.5px_var(--c-subtle)]",
  offline: "bg-(--dot-ring,var(--c-surface)) shadow-[inset_0_0_0_1.5px_var(--c-subtle)]",
  online: "bg-success",
};

// 画像がなければ名前の頭文字を、名前から決めた色相の背景に表示する。
// 状態の点の縁は --dot-ring で背景色に合わせる（サイドバーなど）
export const Avatar = ({ name, src, size = 32, presence }: AvatarProps) => {
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
      {presence && (
        <span
          data-presence={presence}
          className={cn(
            "absolute -right-0.5 -bottom-0.5 rounded-full border-2 border-(--dot-ring,var(--c-surface))",
            presenceClassNames[presence],
          )}
          style={{ height: dotSize, width: dotSize }}
        />
      )}
    </span>
  );
};
