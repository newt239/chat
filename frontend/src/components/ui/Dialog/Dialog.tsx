import type { ReactNode } from "react";

import { IconX } from "@tabler/icons-react";
import { Heading } from "react-aria-components";
import { useTranslation } from "react-i18next";

import { IconButton } from "#/components/ui/IconButton/IconButton";
import { useIsMobile } from "#/hooks/useMediaQuery";

import { DialogFrame } from "./DialogFrame";

const widths = {
  md: "md:max-w-150",
  sm: "md:max-w-110",
};

type DialogProps = {
  isOpen: boolean;
  onOpenChange: (isOpen: boolean) => void;
  title: string;
  children: ReactNode;
  // ボタンを並べる。右寄せで表示する
  footer?: ReactNode;
  size?: keyof typeof widths;
};

// デスクトップでは中央のモーダル、モバイルでは下から出る全画面ページとして同じ内容を描く
export const Dialog = ({
  isOpen,
  onOpenChange,
  title,
  children,
  footer,
  size = "sm",
}: DialogProps) => {
  const { t } = useTranslation();
  const isMobile = useIsMobile();
  return (
    <DialogFrame
      isOpen={isOpen}
      onOpenChange={onOpenChange}
      layout={isMobile ? "sheet" : "center"}
      role="dialog"
      className={widths[size]}
    >
      <header className="flex items-center gap-2 pt-3 pr-2.5 pl-4.5 md:pt-4">
        <Heading slot="title" className="m-0 min-w-0 flex-1 truncate text-base font-bold">
          {title}
        </Heading>
        <IconButton
          label={t("common.close")}
          onPress={() => {
            onOpenChange(false);
          }}
        >
          <IconX />
        </IconButton>
      </header>
      <div className="flex min-h-0 flex-1 flex-col gap-3 overflow-y-auto px-4.5 pt-2.5 pb-1 text-body-sm">
        {children}
      </div>
      {footer && (
        <footer className="flex justify-end gap-2 px-4.5 pt-3.5 pb-[max(16px,env(safe-area-inset-bottom))]">
          {footer}
        </footer>
      )}
    </DialogFrame>
  );
};
