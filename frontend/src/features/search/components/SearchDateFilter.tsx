import { IconChevronDown } from "@tabler/icons-react";
import { Button as AriaButton, DialogTrigger } from "react-aria-components";
import { useTranslation } from "react-i18next";

import { Button } from "#/components/ui/Button";
import { Popover } from "#/components/ui/Popover";
import { TextField } from "#/components/ui/TextField";
import { chipClassName } from "#/features/search/utils/chipClassName";

type DateRange = { after: string | null; before: string | null };

type SearchDateFilterProps = DateRange & {
  isInvalid: boolean;
  onChange: (range: DateRange) => void;
};

// 期間のチップ。開始日・終了日を after: / before: として扱う
export const SearchDateFilter = ({ after, before, isInvalid, onChange }: SearchDateFilterProps) => {
  const { t } = useTranslation();
  const label = t("search.filters.date");
  const isActive = after !== null || before !== null;
  return (
    <DialogTrigger>
      <AriaButton data-active={isActive} data-invalid={isInvalid} className={chipClassName}>
        {isActive ? `${label}: ${after ?? ""}〜${before ?? ""}` : label}
        <IconChevronDown aria-hidden />
      </AriaButton>
      <Popover aria-label={label} placement="bottom start" className="w-64 p-3">
        <div className="flex flex-col gap-3">
          <TextField
            type="date"
            label={t("search.filters.dateFrom")}
            value={after ?? ""}
            onChange={(value) => {
              onChange({ after: value || null, before });
            }}
          />
          <TextField
            type="date"
            label={t("search.filters.dateTo")}
            value={before ?? ""}
            onChange={(value) => {
              onChange({ after, before: value || null });
            }}
          />
          <Button
            variant="secondary"
            size="sm"
            isDisabled={!isActive}
            onPress={() => {
              onChange({ after: null, before: null });
            }}
          >
            {t("search.filters.dateClear")}
          </Button>
        </div>
      </Popover>
    </DialogTrigger>
  );
};
