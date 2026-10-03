import { IconX } from "@tabler/icons-react";
import { useAtom } from "jotai";
import { AnimatePresence, motion } from "motion/react";
import { useTranslation } from "react-i18next";

import { IconButton } from "#/components/ui/IconButton/IconButton";
import { ResizeHandle } from "#/components/ui/ResizeHandle/ResizeHandle";
import { transitions } from "#/lib/motion";
import { sidebarWidthRanges, sidebarWidthsAtom } from "#/providers/store/ui";

import { useRightPanel } from "../hooks/useRightPanel";

type RightSidePanelProps = {
  workspaceId: string;
};

// スレッド・プロフィール・メンバー・チャンネル情報・ピン留めを切り替えて表示する枠
export const RightSidePanel = ({ workspaceId }: RightSidePanelProps) => {
  const { t } = useTranslation();
  const { close, content } = useRightPanel(workspaceId);
  const [widths, setWidths] = useAtom(sidebarWidthsAtom);

  return (
    <AnimatePresence initial={false}>
      {content && (
        <motion.aside
          key="right-panel"
          aria-label={content.title}
          initial={{ width: 0 }}
          // 開閉だけアニメーションし、幅の変更には内側の width がそのまま追従する
          animate={{ width: "auto" }}
          exit={{ width: 0 }}
          transition={transitions.base}
          className="relative flex shrink-0 overflow-clip border-l border-border bg-surface"
        >
          <ResizeHandle
            label={t("shell.rightPanel.resize")}
            edge="left"
            value={widths.right}
            {...sidebarWidthRanges.right}
            onChange={(right) => {
              setWidths({ ...widths, right });
            }}
          />
          <div style={{ width: widths.right }} className="flex shrink-0 flex-col">
            <header className="flex h-12 shrink-0 items-center gap-1 border-b border-border pr-2 pl-4">
              <h2 className="m-0 min-w-0 flex-1 truncate text-body font-bold">{content.title}</h2>
              {content.extra}
              <IconButton label={t("common.close")} onPress={close}>
                <IconX />
              </IconButton>
            </header>
            <motion.div
              key={content.key}
              initial={{ opacity: 0, x: 14 }}
              animate={{ opacity: 1, x: 0 }}
              transition={transitions.spring}
              className="flex min-h-0 flex-1 flex-col overflow-y-auto"
            >
              {content.body}
            </motion.div>
          </div>
        </motion.aside>
      )}
    </AnimatePresence>
  );
};
