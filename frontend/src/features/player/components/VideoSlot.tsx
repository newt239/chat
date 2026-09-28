import { useLayoutEffect, useRef } from "react";

import { mediaPlayer } from "../mediaPlayer";

type VideoSlotProps = {
  kind: "inline" | "mini";
};

// 共有の <video> を差し込む枠。React はこの要素の子を描画しない
export const VideoSlot = ({ kind }: VideoSlotProps) => {
  const ref = useRef<HTMLDivElement>(null);
  useLayoutEffect(() => {
    if (ref.current === null) {
      return undefined;
    }
    return mediaPlayer.registerSlot(kind, ref.current);
  }, [kind]);
  return <div ref={ref} className="absolute inset-0 bg-media" />;
};
