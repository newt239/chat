import { useLayoutEffect, useRef } from "react";
import type { ReactNode } from "react";

import { animate, motion, useDragControls, useMotionValue } from "motion/react";

import { transitions } from "#/lib/motion";

import { MobileBackContext } from "../hooks/useMobileBack";
import { MobileForwardContext } from "../hooks/useMobileForward";
import { isSwipeBlocked } from "../utils/swipe";

type MobileStackLayerProps = {
  // 画面を閉じ終えたあとに呼ぶ（履歴を戻る、パネルを閉じるなど）
  onBack: () => void;
  children: ReactNode;
};

// 右から重なる画面。右へのスワイプか「戻る」で右へ退いてから onBack を呼ぶ。左へのスワイプは中身が登録した操作（チャンネル情報など）を開く
export const MobileStackLayer = ({ onBack, children }: MobileStackLayerProps) => {
  const ref = useRef<HTMLDivElement>(null);
  const x = useMotionValue(0);
  const dragControls = useDragControls();
  const forwardRef = useRef<(() => void) | null>(null);
  const width = () => ref.current?.offsetWidth ?? 0;

  useLayoutEffect(() => {
    x.set(width());
    const controls = animate(x, 0, transitions.push);
    return () => {
      controls.stop();
    };
  }, [x]);

  const back = () => {
    void animate(x, width(), transitions.push).then(onBack);
  };

  return (
    <MobileBackContext value={back}>
      <MobileForwardContext value={forwardRef}>
        <motion.div
          ref={ref}
          style={{ x }}
          drag="x"
          dragListener={false}
          dragControls={dragControls}
          dragDirectionLock
          dragConstraints={{ left: 0, right: 0 }}
          dragElastic={{ left: 0.2, right: 1 }}
          onPointerDown={(event) => {
            if (
              ref.current !== null &&
              event.target instanceof Element &&
              !isSwipeBlocked(event.target, ref.current)
            ) {
              dragControls.start(event);
            }
          }}
          onDragEnd={(_, info) => {
            if (info.offset.x > width() / 3 || info.velocity.x > 500) {
              back();
            } else if (info.offset.x < -width() / 4 || info.velocity.x < -500) {
              forwardRef.current?.();
            }
          }}
          // 入力欄をホームインジケーターから離す。キーボードが出ている間は不要
          className="absolute inset-0 isolate flex touch-pan-y flex-col overscroll-contain bg-surface pb-[env(safe-area-inset-bottom)] shadow-xl group-data-keyboard/shell:pb-0"
        >
          {children}
        </motion.div>
      </MobileForwardContext>
    </MobileBackContext>
  );
};
