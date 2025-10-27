import { ActionIcon, Group } from "@mantine/core";
import { IconMenu2, IconInfoCircle, IconBookmark, IconSearch } from "@tabler/icons-react";
import { useAtomValue, useSetAtom } from "jotai";

import {
  isMobileAtom,
  mobileActivePanelAtom,
  showMobileLeftPanelAtom,
  showMobileRightPanelAtom,
  setRightSidePanelViewAtom,
} from "@/providers/store/ui";

export const MobileBottomBar = () => {
  const isMobile = useAtomValue(isMobileAtom);
  const mobileActivePanel = useAtomValue(mobileActivePanelAtom);
  const showMobileLeftPanel = useSetAtom(showMobileLeftPanelAtom);
  const showMobileRightPanel = useSetAtom(showMobileRightPanelAtom);
  const setRightSidePanelView = useSetAtom(setRightSidePanelViewAtom);

  // モバイルでない場合は非表示
  if (!isMobile) {
    return null;
  }

  const handleLeftPanelClick = () => {
    if (mobileActivePanel === "left") {
      // 既に左パネルが表示されている場合は閉じる
      showMobileLeftPanel();
    } else {
      showMobileLeftPanel();
    }
  };

  const handleRightPanelClick = () => {
    if (mobileActivePanel === "right") {
      // 既に右パネルが表示されている場合は閉じる
      showMobileRightPanel();
    } else {
      showMobileRightPanel();
    }
  };

  const handleBookmarkClick = () => {
    setRightSidePanelView({ type: "bookmarks" });
    showMobileRightPanel();
  };

  const handleSearchClick = () => {
    // TODO: 検索機能の実装
    console.log("Search clicked");
  };

  return (
    <div className="fixed bottom-0 left-0 right-0 bg-white border-t border-gray-200 px-4 py-2 z-50">
      <Group justify="space-around" gap="xs">
        {/* チャンネル一覧ボタン */}
        <ActionIcon
          variant={mobileActivePanel === "left" ? "filled" : "subtle"}
          size="lg"
          onClick={handleLeftPanelClick}
          className={`${
            mobileActivePanel === "left"
              ? "bg-blue-600 text-white"
              : "text-gray-700 hover:bg-gray-100"
          }`}
          title="チャンネル一覧"
        >
          <IconMenu2 size={20} />
        </ActionIcon>

        {/* 検索ボタン */}
        <ActionIcon
          variant="subtle"
          size="lg"
          onClick={handleSearchClick}
          className="text-gray-700 hover:bg-gray-100"
          title="検索"
        >
          <IconSearch size={20} />
        </ActionIcon>

        {/* ブックマークボタン */}
        <ActionIcon
          variant="subtle"
          size="lg"
          onClick={handleBookmarkClick}
          className="text-gray-700 hover:bg-gray-100"
          title="ブックマーク"
        >
          <IconBookmark size={20} />
        </ActionIcon>

        {/* チャンネル情報ボタン */}
        <ActionIcon
          variant={mobileActivePanel === "right" ? "filled" : "subtle"}
          size="lg"
          onClick={handleRightPanelClick}
          className={`${
            mobileActivePanel === "right"
              ? "bg-blue-600 text-white"
              : "text-gray-700 hover:bg-gray-100"
          }`}
          title="チャンネル情報"
        >
          <IconInfoCircle size={20} />
        </ActionIcon>
      </Group>
    </div>
  );
};
