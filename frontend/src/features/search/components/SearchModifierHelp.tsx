import { Button } from "react-aria-components";
import { useTranslation } from "react-i18next";

import { focusRing } from "#/components/ui/styles";
import { Tooltip } from "#/components/ui/Tooltip";

const modifiers = [
  ["from:@", "from"],
  ["in:#", "in"],
  ["has:image", "has"],
  ["is:pinned", "pinned"],
  ["is:thread", "thread"],
  ["is:mention", "mention"],
  ["after:", "after"],
  ["before:", "before"],
] as const;

type SearchModifierHelpProps = {
  onInsert: (modifier: string) => void;
};

// 入力欄の下に並べる修飾子の見本。押すと入力欄に挿入する。モバイルでは条件のチップで足りるため出さない
export const SearchModifierHelp = ({ onInsert }: SearchModifierHelpProps) => {
  const { t } = useTranslation();
  return (
    <div className="flex flex-wrap items-center max-md:hidden gap-x-2.5 gap-y-1 text-[11.5px] text-subtle">
      {modifiers.map(([modifier, key]) => (
        <Tooltip key={modifier} content={t(`search.help.${key}`)}>
          <Button
            aria-label={t("search.help.insert", { modifier })}
            onPress={() => {
              onInsert(modifier);
            }}
            className={`cursor-pointer rounded-[4px] border border-border bg-sunken px-1 font-mono text-[11px] text-muted data-hovered:text-text ${focusRing}`}
          >
            {modifier}
          </Button>
        </Tooltip>
      ))}
      <span>{t("search.help.hint")}</span>
    </div>
  );
};
