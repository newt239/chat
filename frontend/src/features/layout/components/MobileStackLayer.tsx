import { useLayoutEffect, useRef } from "react";
import type { ReactNode } from "react";

import { animate, motion, useDragControls, useMotionValue } from "motion/react";

import { transitions } from "#/lib/motion";

import { MobileBackContext } from "../hooks/useMobileBack";

type MobileStackLayerProps = {
  // 画面を閉じ終えたあとに呼ぶ（履歴を戻る、パネルを閉じるなど）
  onBack: () => void;
  children: ReactNode;
};

// 右から重なる画面。左端からのスワイプか「戻る」で右へ退いてから onBack を呼ぶ
export const MobileStackLayer = ({ onBack, children }: MobileStackLayerProps) => {
  const ref = useRef<HTMLDivElement>(null);
  const x = useMotionValue(0);
  const dragControls = useDragControls();
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
      <motion.div
        ref={ref}
        style={{ x }}
        drag="x"
        dragListener={false}
        dragControls={dragControls}
        dragConstraints={{ left: 0, right: 0 }}
        dragElastic={{ left: 0, right: 1 }}
        onDragEnd={(_, info) => {
          if (info.offset.x > width() / 3 || info.velocity.x > 500) {
            back();
          }
        }}
        // 入力欄をホームインジケーターから離す。キーボードが出ている間は不要
        className="absolute inset-0 flex flex-col bg-surface pb-[env(safe-area-inset-bottom)] shadow-xl group-data-keyboard/shell:pb-0"
      >
        <div
          aria-hidden
          className="absolute inset-y-0 left-0 z-10 w-4 touch-none"
          onPointerDown={(event) => {
            dragControls.start(event);
          }}
        />
        {children}
      </motion.div>
    </MobileBackContext>
  );
};
