import { IconPlayerPauseFilled, IconPlayerPlayFilled } from "@tabler/icons-react";
import { Button } from "react-aria-components";
import { useTranslation } from "react-i18next";

import { cn, focusRing } from "#/components/ui/styles";

type PlayPauseButtonProps = {
  isPlaying: boolean;
  onPress: () => void;
  // 大きさと色。置く場所ごとに違う
  className: string;
};

export const PlayPauseButton = ({ isPlaying, onPress, className }: PlayPauseButtonProps) => {
  const { t } = useTranslation();
  return (
    <Button
      aria-label={t(isPlaying ? "attachment.player.pause" : "attachment.player.play")}
      onPress={onPress}
      className={cn("grid shrink-0 cursor-pointer place-items-center", className, focusRing)}
    >
      {isPlaying ? <IconPlayerPauseFilled aria-hidden /> : <IconPlayerPlayFilled aria-hidden />}
    </Button>
  );
};
