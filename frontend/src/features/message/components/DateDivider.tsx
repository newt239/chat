import { useRef, useState } from "react";

import { IconChevronDown } from "@tabler/icons-react";
import { useNavigate } from "@tanstack/react-router";
import { Button } from "react-aria-components";
import { useTranslation } from "react-i18next";

import { Calendar } from "#/components/ui/Calendar/Calendar";
import { Menu } from "#/components/ui/Menu/Menu";
import { MenuItem } from "#/components/ui/MenuItem/MenuItem";
import { MenuSeparator } from "#/components/ui/MenuSeparator/MenuSeparator";
import { Popover } from "#/components/ui/Popover/Popover";
import { focusRing } from "#/components/ui/styles/styles";
import { useDateFormat } from "#/hooks/useDateFormat";

import { FIRST_MESSAGE, jumpPresets, startOfDateKey } from "../utils/dateJump";

type DateDividerProps = {
  dateKey: string;
  // 一覧の上端に重ねて、表示中の日付を示す。線は出さない
  floating: boolean;
};

// その日の投稿の上に置く区切り。押すと別の日へ移動できる
export const DateDivider = ({ dateKey, floating }: DateDividerProps) => {
  const { t } = useTranslation();
  const { formatDateWithWeekday, timeZone } = useDateFormat();
  const navigate = useNavigate();
  const anchorRef = useRef<HTMLDivElement>(null);
  const [isPickerOpen, setIsPickerOpen] = useState(false);
  const presets = jumpPresets(new Date(), timeZone);

  const jump = (date: string) => {
    void navigate({ search: (prev) => ({ ...prev, date, message: undefined }), to: "." });
  };

  const label =
    dateKey === presets.today
      ? t("message.date.today")
      : dateKey === presets.yesterday
        ? t("message.date.yesterday")
        : formatDateWithWeekday(startOfDateKey(dateKey, timeZone));

  const pill = (
    <div
      ref={anchorRef}
      className={
        floating
          ? "pointer-events-none absolute inset-x-0 top-1.5 z-10 flex justify-center font-sans"
          : "pointer-events-none relative z-10 -mt-2.75 mb-1 flex justify-center font-sans"
      }
    >
      <Menu
        placement="bottom"
        trigger={
          <Button
            aria-label={`${label} · ${t("message.date.jumpTo")}`}
            className={`pointer-events-auto inline-flex cursor-pointer items-center gap-1 rounded-full border border-border bg-surface py-0.5 pr-2 pl-3 text-caption font-semibold text-text shadow-sm data-hovered:bg-hover data-pressed:bg-hover [&_svg]:size-3.5 [&_svg]:text-muted ${focusRing}`}
          >
            {label}
            <IconChevronDown aria-hidden />
          </Button>
        }
      >
        {(["today", "yesterday", "lastWeek", "lastMonth"] as const).map((preset) => (
          <MenuItem
            key={preset}
            onAction={() => {
              jump(presets[preset]);
            }}
          >
            {t(`message.date.${preset}`)}
          </MenuItem>
        ))}
        <MenuItem
          onAction={() => {
            jump(FIRST_MESSAGE);
          }}
        >
          {t("message.date.first")}
        </MenuItem>
        <MenuSeparator />
        <MenuItem
          onAction={() => {
            setIsPickerOpen(true);
          }}
        >
          {t("message.date.pick")}
        </MenuItem>
      </Menu>
      <Popover
        aria-label={t("message.date.calendar")}
        triggerRef={anchorRef}
        isOpen={isPickerOpen}
        onOpenChange={setIsPickerOpen}
        placement="bottom"
      >
        <Calendar
          aria-label={t("message.date.calendar")}
          maxDate={presets.today}
          onChange={(date) => {
            setIsPickerOpen(false);
            jump(date);
          }}
        />
      </Popover>
    </div>
  );

  if (floating) {
    return pill;
  }
  return (
    <>
      <div aria-hidden className="mt-3 border-t border-border" />
      {pill}
    </>
  );
};
