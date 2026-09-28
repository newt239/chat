import { ToggleButton } from "react-aria-components";
import { useTranslation } from "react-i18next";

import { focusRing } from "#/components/ui/styles";
import { Tooltip } from "#/components/ui/Tooltip";

type DescendantsToggleProps = {
  count: number;
  isSelected: boolean;
  onChange: (isSelected: boolean) => void;
};

// 親チャンネルで子孫チャンネルのメッセージもまとめて表示するかの切り替え
export const DescendantsToggle = ({ count, isSelected, onChange }: DescendantsToggleProps) => {
  const { t } = useTranslation();
  return (
    <Tooltip content={t("channel.aggregate.hint", { count })}>
      <ToggleButton
        isSelected={isSelected}
        onChange={onChange}
        className={`group inline-flex shrink-0 cursor-pointer items-center gap-1.5 rounded-full border border-border py-px pr-2.5 pl-[5px] text-[11.5px] whitespace-nowrap text-muted data-selected:border-accent data-selected:bg-accent-soft data-selected:text-accent-text ${focusRing}`}
      >
        <span
          aria-hidden
          className="relative block h-[13px] w-[22px] rounded-full bg-border-strong group-data-selected:bg-accent"
        >
          <span className="absolute top-0.5 left-0.5 size-[9px] rounded-full bg-surface transition-transform group-data-selected:translate-x-[9px] motion-reduce:transition-none" />
        </span>
        {t("channel.aggregate.toggle")}
      </ToggleButton>
    </Tooltip>
  );
};
