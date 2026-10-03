import { IconChevronLeft, IconChevronRight } from "@tabler/icons-react";
import {
  Calendar as AriaCalendar,
  CalendarCell,
  CalendarGrid,
  CalendarGridBody,
  CalendarGridHeader,
  CalendarHeaderCell,
  Heading,
} from "react-aria-components";
import { useTranslation } from "react-i18next";

import { IconButton } from "#/components/ui/IconButton/IconButton";
import { focusRing } from "#/components/ui/styles/styles";

type CalendarProps = {
  "aria-label": string;
  // YYYY-MM-DD。これより後の日は選べない
  maxDate: string;
  onChange: (date: string) => void;
};

// 日付を YYYY-MM-DD の文字列でやり取りするカレンダー
export const Calendar = ({ "aria-label": ariaLabel, maxDate, onChange }: CalendarProps) => {
  const { t } = useTranslation();
  return (
    <AriaCalendar
      aria-label={ariaLabel}
      // YYYY-MM-DD どうしは文字列の大小で日付の前後を比べられる
      isDateUnavailable={(date) => date.toString() > maxDate}
      onChange={(date) => {
        onChange(date.toString());
      }}
      className="flex flex-col gap-2 p-3 font-sans"
    >
      <header className="flex items-center gap-1">
        <IconButton slot="previous" label={t("ui.calendar.previous")}>
          <IconChevronLeft aria-hidden />
        </IconButton>
        <Heading className="m-0 flex-1 text-center text-body-strong" />
        <IconButton slot="next" label={t("ui.calendar.next")}>
          <IconChevronRight aria-hidden />
        </IconButton>
      </header>
      <CalendarGrid className="border-collapse">
        <CalendarGridHeader>
          {(day) => (
            <CalendarHeaderCell className="pb-1 text-caption font-normal text-muted">
              {day}
            </CalendarHeaderCell>
          )}
        </CalendarGridHeader>
        <CalendarGridBody>
          {(date) => (
            <CalendarCell
              date={date}
              className={`grid size-8 cursor-pointer place-items-center rounded-md text-body-sm text-text tabular-nums data-hovered:bg-hover data-outside-month:hidden data-selected:bg-accent data-selected:text-accent-fg data-unavailable:cursor-default data-unavailable:text-subtle data-unavailable:line-through ${focusRing}`}
            />
          )}
        </CalendarGridBody>
      </CalendarGrid>
    </AriaCalendar>
  );
};
