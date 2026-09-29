import type { ReactNode } from "react";

import { AnimatePresence, motion } from "motion/react";
import { Dialog as AriaDialog, Modal, ModalOverlay } from "react-aria-components";

import { cn } from "#/components/ui/styles/styles";
import { transitions } from "#/lib/motion";

const MotionModalOverlay = motion.create(ModalOverlay);
const MotionModal = motion.create(Modal);

type DialogFrameProps = {
  isOpen: boolean;
  onOpenChange: (isOpen: boolean) => void;
  // sheet: 下から出る全画面（モバイル）、bottom: 下から出る高さ可変のシート、center: 中央のモーダル
  layout: keyof typeof layouts;
  role: "dialog" | "alertdialog";
  className: string;
  children: ReactNode;
};

const layouts = {
  bottom: {
    animation: { animate: { y: 0 }, exit: { y: "100%" }, initial: { y: "100%" } },
    className:
      "fixed inset-x-0 bottom-0 max-h-[88%] overflow-y-auto rounded-t-[18px] pb-[max(24px,env(safe-area-inset-bottom))]",
    transition: transitions.sheet,
  },
  center: {
    animation: {
      animate: { opacity: 1, scale: 1 },
      exit: { opacity: 0, scale: 0.96 },
      initial: { opacity: 0, scale: 0.96 },
    },
    className: "max-h-full w-full rounded-xl shadow-xl",
    transition: transitions.fast,
  },
  sheet: {
    animation: { animate: { y: 0 }, exit: { y: "100%" }, initial: { y: "100%" } },
    className: "fixed inset-0 rounded-none pt-[env(safe-area-inset-top)]",
    transition: transitions.sheet,
  },
};

// Dialog と AlertDialog が共有するオーバーレイとアニメーション
export const DialogFrame = ({
  isOpen,
  onOpenChange,
  layout,
  role,
  className,
  children,
}: DialogFrameProps) => {
  const { animation, className: layoutClassName, transition } = layouts[layout];
  return (
    <AnimatePresence>
      {isOpen && (
        <MotionModalOverlay
          isOpen
          onOpenChange={onOpenChange}
          isDismissable={role === "dialog"}
          initial={{ opacity: 0 }}
          animate={{ opacity: 1 }}
          exit={{ opacity: 0 }}
          transition={transitions.fast}
          className="fixed inset-0 z-[300] grid place-items-center bg-overlay px-4 py-6"
        >
          <MotionModal
            {...animation}
            transition={transition}
            className={cn(
              "flex flex-col bg-surface font-sans text-text",
              layoutClassName,
              className,
            )}
          >
            <AriaDialog role={role} className="flex min-h-0 flex-1 flex-col outline-none">
              {children}
            </AriaDialog>
          </MotionModal>
        </MotionModalOverlay>
      )}
    </AnimatePresence>
  );
};
