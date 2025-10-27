import { useAtomValue } from "jotai";

import { CenterPanel } from "./CenterPanel";
import { LeftSidePanel } from "./LeftSidePanel";
import { MobileBottomBar } from "./MobileBottomBar";
import { RightSidePanel } from "./RightSidePanel";

import { GlobalHeaderPanel } from "@/features/workspace/components/Header";
import {
  leftSidePanelVisibleAtom,
  rightSidePanelViewAtom,
  mobileActivePanelAtom,
  isChannelPageAtom,
} from "@/providers/store/ui";

export const ResponsiveLayout = () => {
  const leftSidePanelVisible = useAtomValue(leftSidePanelVisibleAtom);
  const rightSidePanelView = useAtomValue(rightSidePanelViewAtom);
  const mobileActivePanel = useAtomValue(mobileActivePanelAtom);
  const isChannelPage = useAtomValue(isChannelPageAtom);

  return (
    <div className="h-full flex flex-col bg-gray-50">
      {/* グローバルヘッダー（デスクトップのみ表示） */}
      <div className="hidden md:block">
        <GlobalHeaderPanel />
      </div>

      {/* メインコンテンツエリア */}
      <div className="flex-1 min-h-0 flex relative">
        {/* 左サイドパネル（デスクトップ表示） */}
        {leftSidePanelVisible && (
          <div className="hidden md:block w-80 shrink-0">
            <LeftSidePanel className="h-full" />
          </div>
        )}

        {/* 左サイドパネル（モバイルオーバーレイ） */}
        {mobileActivePanel === "left" && (
          <div className="md:hidden fixed inset-0 z-40 bg-black bg-opacity-50">
            <div className="w-80 h-full">
              <LeftSidePanel className="h-full" />
            </div>
          </div>
        )}

        {/* 中央パネル */}
        <div className="flex-1 min-w-0">
          <CenterPanel />
        </div>

        {/* 右サイドパネル（デスクトップ表示） */}
        {rightSidePanelView.type !== "hidden" && (
          <div className="hidden md:block w-80 shrink-0">
            <RightSidePanel className="h-full" />
          </div>
        )}

        {/* 右サイドパネル（モバイルオーバーレイ） */}
        {mobileActivePanel === "right" && rightSidePanelView.type !== "hidden" && (
          <div className="md:hidden fixed inset-0 z-40 bg-black bg-opacity-50">
            <div className="w-80 h-full ml-auto">
              <RightSidePanel className="h-full" />
            </div>
          </div>
        )}
      </div>

      {/* モバイルボトムバー（チャンネルページでない場合のみ表示） */}
      {!isChannelPage && (
        <div className="md:hidden">
          <MobileBottomBar />
        </div>
      )}
    </div>
  );
};
