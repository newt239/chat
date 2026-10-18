import { buildTokens } from "@chat/design-tokens/theme";

import { useColorMode } from "#/providers/theme/colorMode";

import type { ThemeInput } from "@chat/design-tokens/theme";

type ThemePreviewProps = {
  theme: ThemeInput;
};

// 地色の上にパネルが並ぶ配色を縮小して見せる。色はテーマから計算した値をそのまま使う
export const ThemePreview = ({ theme }: ThemePreviewProps) => {
  const tokens = buildTokens(theme, useColorMode());
  return (
    <span
      aria-hidden
      className="flex h-16 gap-1 overflow-hidden p-1"
      style={{ background: tokens.bg }}
    >
      <span
        className="flex w-1/3 flex-col gap-1.5 rounded-sm p-1.5"
        style={{ background: tokens.surface }}
      >
        {[70, 90, 60].map((width, index) => (
          <i
            key={width}
            className="block h-1.5 rounded-full"
            style={{
              background: index === 1 ? tokens.accent : tokens.muted,
              opacity: index === 1 ? 1 : 0.5,
              width: `${width}%`,
            }}
          />
        ))}
      </span>
      <span
        className="flex flex-1 flex-col gap-1.5 rounded-sm p-1.5"
        style={{ background: tokens.surface }}
      >
        <i className="block h-1.5 w-1/2 rounded-full" style={{ background: tokens.text }} />
        <i className="block h-1.5 w-4/5 rounded-full" style={{ background: tokens.muted }} />
        <i
          className="mt-auto block h-2.5 w-2/5 self-end rounded-sm"
          style={{ background: tokens.accent }}
        />
      </span>
    </span>
  );
};
