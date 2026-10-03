import { useState } from "react";

import { IconCalendarEvent, IconChevronDown, IconClock } from "@tabler/icons-react";
import { useTranslation } from "react-i18next";

import { IconButton } from "#/components/ui/IconButton/IconButton";
import { Menu } from "#/components/ui/Menu/Menu";
import { MenuItem } from "#/components/ui/MenuItem/MenuItem";
import { MenuSeparator } from "#/components/ui/MenuSeparator/MenuSeparator";
import { useDateFormat } from "#/hooks/useDateFormat";

import { schedulePresets } from "../utils/schedulePresets";
import { ScheduleDialog } from "./ScheduleDialog";

type ScheduleSendMenuProps = {
  isDisabled: boolean;
  onSchedule: (scheduledAt: Date) => void;
};

// 送信ボタンの横のメニュー。よく使う日時か、任意の日時で送信を予約する
export const ScheduleSendMenu = ({ isDisabled, onSchedule }: ScheduleSendMenuProps) => {
  const { t } = useTranslation();
  const [isDialogOpen, setIsDialogOpen] = useState(false);
  const { timeZone } = useDateFormat();
  const presets = schedulePresets(new Date(), timeZone);

  return (
    <>
      <Menu
        placement="top end"
        trigger={
          <IconButton
            label={t("schedule.menu.label")}
            isDisabled={isDisabled}
            className="h-7 w-6 [&_svg]:size-3.5"
          >
            <IconChevronDown />
          </IconButton>
        }
      >
        {presets.map(({ key, date }) => (
          <MenuItem
            key={key}
            icon={<IconClock aria-hidden />}
            onAction={() => {
              onSchedule(date);
            }}
          >
            {t(`schedule.presets.${key}`)}
          </MenuItem>
        ))}
        <MenuSeparator />
        <MenuItem
          icon={<IconCalendarEvent aria-hidden />}
          onAction={() => {
            setIsDialogOpen(true);
          }}
        >
          {t("schedule.menu.custom")}
        </MenuItem>
      </Menu>
      {isDialogOpen && (
        <ScheduleDialog
          onClose={() => {
            setIsDialogOpen(false);
          }}
          title={t("schedule.dialog.title")}
          // 日時を指定するときは明日の朝を初期値にする
          initialDate={presets[1].date}
          isPending={false}
          onConfirm={(scheduledAt) => {
            setIsDialogOpen(false);
            onSchedule(scheduledAt);
          }}
        />
      )}
    </>
  );
};
