import { IconLoader2, IconPlayerStopFilled, IconX } from "@tabler/icons-react";
import { useTranslation } from "react-i18next";

import { Button } from "#/components/ui/Button/Button";
import { IconButton } from "#/components/ui/IconButton/IconButton";
import { formatDuration } from "#/lib/formatDuration";

import { useVoiceRecorder } from "../hooks/useVoiceRecorder";

type VoiceRecorderProps = {
  // 録音時間は MediaRecorder の webm から読めないことがあるため、計測した値を一緒に渡す
  onAttach: (file: File, durationSeconds: number) => void;
  onDiscard: () => void;
};

// 入力欄の中で録音し、止めたら試聴してから添付するか破棄する
export const VoiceRecorder = ({ onAttach, onDiscard }: VoiceRecorderProps) => {
  const { t } = useTranslation();
  const { state, stop } = useVoiceRecorder();

  return (
    <div className="mx-2 mt-2 flex min-h-10 items-center gap-2 rounded-md border border-border bg-sunken py-1 pr-1 pl-2.5 font-sans text-caption text-muted">
      {state.status === "starting" && (
        <span className="inline-flex flex-1 items-center gap-1.5 [&_svg]:size-4">
          <IconLoader2 aria-hidden className="animate-spin motion-reduce:animate-none" />
          {t("recorder.preparing")}
        </span>
      )}
      {state.status === "failed" && (
        <span role="alert" className="flex-1 text-danger">
          {t(`recorder.failed.${state.reason}`)}
        </span>
      )}
      {state.status === "recording" && (
        <>
          <span className="inline-flex flex-1 items-center gap-2 text-text">
            <span
              aria-hidden
              className="size-2.5 rounded-full bg-danger motion-safe:animate-pulse"
            />
            {t("recorder.recording")}
            <span className="font-mono tabular-nums">{formatDuration(state.elapsedSeconds)}</span>
          </span>
          <Button variant="secondary" size="sm" onPress={stop}>
            <IconPlayerStopFilled aria-hidden />
            {t("recorder.stop")}
          </Button>
        </>
      )}
      {state.status === "recorded" && (
        <>
          {/* oxlint-disable-next-line jsx-a11y/media-has-caption -- 自分の録音を聴き直すだけで字幕はない */}
          <audio controls src={state.url} className="h-8 min-w-0 flex-1" />
          <Button
            size="sm"
            onPress={() => {
              onAttach(state.file, state.durationSeconds);
            }}
          >
            {t("recorder.attach")}
          </Button>
        </>
      )}
      <IconButton
        label={t("recorder.discard")}
        onPress={onDiscard}
        className="size-7 [&_svg]:size-4"
      >
        <IconX />
      </IconButton>
    </div>
  );
};
