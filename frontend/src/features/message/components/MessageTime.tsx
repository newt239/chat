import { formatFullDateTime, formatTime } from "@chat/i18n";
import { useAtomValue } from "jotai";
import { Focusable } from "react-aria-components";

import { focusRing } from "#/components/ui/styles";
import { Tooltip } from "#/components/ui/Tooltip";
import { preferencesAtom } from "#/providers/store/preferences";

type MessageTimeProps = {
  date: Date;
};

// 時刻だけを表示し、ホバー・フォーカスで曜日と秒を含む日時を出す
export const MessageTime = ({ date }: MessageTimeProps) => {
  const { locale } = useAtomValue(preferencesAtom);
  const time = formatTime(date, locale);

  return (
    <Tooltip content={formatFullDateTime(date, locale)}>
      <Focusable>
        {/* ツールチップの aria-describedby が読み上げられる役割のうち、操作を持たない img にする */}
        <time
          role="img"
          aria-label={time}
          // oxlint-disable-next-line jsx-a11y/no-noninteractive-tabindex -- キーボードでもツールチップを開けるようにする
          tabIndex={0}
          dateTime={date.toISOString()}
          className={`shrink-0 rounded-sm font-mono text-[11.5px] text-subtle tabular-nums ${focusRing}`}
        >
          {time}
        </time>
      </Focusable>
    </Tooltip>
  );
};
