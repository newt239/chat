import { Button } from "react-aria-components";
import { useTranslation } from "react-i18next";

import { focusRing } from "#/components/ui/styles";

type SpeedButtonProps = {
  rate: number;
  onPress: () => void;
};

// 押すたびに 1x → 1.5x → 2x と切り替える
export const SpeedButton = ({ rate, onPress }: SpeedButtonProps) => {
  const { t } = useTranslation();
  return (
    <Button
      aria-label={t("attachment.player.speed")}
      onPress={onPress}
      className={`shrink-0 cursor-pointer rounded-[5px] border border-current/40 px-1.5 py-0.5 text-current opacity-80 data-hovered:opacity-100 ${focusRing}`}
    >
      <span className="font-mono text-[11px] font-semibold">{rate}x</span>
    </Button>
  );
};
