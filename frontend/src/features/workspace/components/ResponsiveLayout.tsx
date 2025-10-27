import { useEffect } from "react";

import { useAtomValue, useSetAtom } from "jotai";

import { CenterPanel } from "./CenterPanel";
import { GlobalHeaderPanel } from "./Header";
import { LeftSidePanel } from "./LeftSidePanel";
import { MobileBottomBar } from "./MobileBottomBar";
import { RightSidePanel } from "./RightSidePanel";

import {
  isMobileAtom,
  setIsMobileAtom,
  leftSidePanelVisibleAtom,
  rightSidePanelViewAtom,
  mobileActivePanelAtom,
} from "@/providers/store/ui";

export const ResponsiveLayout = () => {
  const isMobile = useAtomValue(isMobileAtom);
  const setIsMobile = useSetAtom(setIsMobileAtom);
  const leftSidePanelVisible = useAtomValue(leftSidePanelVisibleAtom);
  const rightSidePanelView = useAtomValue(rightSidePanelViewAtom);
  const mobileActivePanel = useAtomValue(mobileActivePanelAtom);

  // 画面サイズの変更を監視してモバイル状態を更新
  useEffect(() => {
    const handleResize = () => {
      const mobile = window.innerWidth < 768; // md breakpoint
      setIsMobile(mobile);
    };

    // 初期設定
    handleResize();

    // リサイズイベントリスナーを追加
    window.addEventListener("resize", handleResize);

    return () => {
      window.removeEventListener("resize", handleResize);
    };
  }, [setIsMobile]);

  // デスクトップレイアウト
  if (!isMobile) {
    return (
      <div className="h-full flex flex-col bg-gray-50">
        {/* グローバルヘッダー */}
        <GlobalHeaderPanel />

        {/* メインコンテンツエリア */}
        <div className="flex-1 min-h-0 flex">
          {/* 左サイドパネル */}
          {leftSidePanelVisible && (
            <div className="w-80 shrink-0">
              <LeftSidePanel className="h-full" />
            </div>
          )}

          {/* 中央パネル */}
          <div className="flex-1 min-w-0">
            <CenterPanel />
          </div>

          {/* 右サイドパネル */}
          {rightSidePanelView.type !== "hidden" && (
            <div className="w-80 shrink-0">
              <RightSidePanel className="h-full" />
            </div>
          )}
        </div>
      </div>
    );
  }

  // モバイルレイアウト
  return (
    <div className="h-full flex flex-col bg-gray-50 relative">
      {/* メインコンテンツエリア */}
      <div className="flex-1 min-h-0 flex relative">
        {/* 左サイドパネル（オーバーレイ） */}
        {mobileActivePanel === "left" && (
          <div className="fixed inset-0 z-40 bg-black bg-opacity-50">
            <div className="w-80 h-full">
              <LeftSidePanel className="h-full" />
            </div>
          </div>
        )}

        {/* 中央パネル */}
        <div className="flex-1 min-w-0">
          <CenterPanel />
        </div>

        {/* 右サイドパネル（オーバーレイ） */}
        {mobileActivePanel === "right" && rightSidePanelView.type !== "hidden" && (
          <div className="fixed inset-0 z-40 bg-black bg-opacity-50">
            <div className="w-80 h-full ml-auto">
              <RightSidePanel className="h-full" />
            </div>
          </div>
        )}
      </div>

      {/* モバイルボトムバー */}
      <MobileBottomBar />
    </div>
  );
};
