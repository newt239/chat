import { useState } from "react";

import { Menu, Button, Text, Avatar, TextInput, ActionIcon } from "@mantine/core";
import { IconSearch, IconBookmark } from "@tabler/icons-react";
import { useNavigate, useParams } from "@tanstack/react-router";
import { useAtomValue, useSetAtom } from "jotai";

import type { WorkspaceSummary } from "@/features/workspace/types";

import { useWorkspaces } from "@/features/workspace/hooks/useWorkspace";
import { clearAuthAtom } from "@/lib/store/auth";
import { currentWorkspaceIdAtom, setCurrentWorkspaceAtom } from "@/lib/store/workspace";
import { setRightSidebarViewAtom } from "@/lib/store/ui";

export const Header = () => {
  const { data: workspaces, isLoading } = useWorkspaces();
  const currentWorkspaceId = useAtomValue(currentWorkspaceIdAtom);
  const setCurrentWorkspace = useSetAtom(setCurrentWorkspaceAtom);
  const clearAuth = useSetAtom(clearAuthAtom);
  const setRightSidebarView = useSetAtom(setRightSidebarViewAtom);
  const [isWorkspaceMenuOpen, setIsWorkspaceMenuOpen] = useState(false);
  const [searchQuery, setSearchQuery] = useState("");
  const navigate = useNavigate();
  const params = useParams({ strict: false });

  const currentWorkspace = workspaces?.find((w: WorkspaceSummary) => w.id === currentWorkspaceId);
  const isInWorkspace = params.workspaceId !== undefined;

  const handleWorkspaceChange = (workspaceId: string) => {
    setCurrentWorkspace(workspaceId);
    setIsWorkspaceMenuOpen(false);
  };

  const handleLogout = () => {
    clearAuth();
  };

  const handleBookmarkClick = () => {
    setRightSidebarView({ type: "bookmarks" });
  };

  const handleSearchSubmit = (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    if (searchQuery.trim() && currentWorkspaceId) {
      navigate({
        to: "/app/$workspaceId/search",
        params: { workspaceId: currentWorkspaceId },
        search: { q: searchQuery.trim(), filter: "all" },
      });
    }
  };

  const handleSearchFocus = () => {
    if (currentWorkspaceId && params.workspaceId !== currentWorkspaceId) {
      navigate({
        to: "/app/$workspaceId/search",
        params: { workspaceId: currentWorkspaceId },
        search: { q: searchQuery.trim() || undefined, filter: "all" },
      });
    }
  };

  return (
    <header className="bg-white border-b border-gray-200 px-4 py-3">
      <div className="flex items-center justify-between gap-4">
        <div className="flex items-center space-x-4 shrink-0">
          {/* ワークスペース選択メニュー */}
          <Menu
            opened={isWorkspaceMenuOpen}
            onClose={() => setIsWorkspaceMenuOpen(false)}
            position="bottom-end"
          >
            <Menu.Target>
              <Button
                variant="subtle"
                onClick={() => setIsWorkspaceMenuOpen(!isWorkspaceMenuOpen)}
                className="text-gray-700 hover:bg-gray-100"
              >
                {isLoading ? (
                  "読み込み中..."
                ) : currentWorkspace ? (
                  <div className="flex items-center space-x-2">
                    <Avatar size="sm" color="blue">
                      {currentWorkspace.name.charAt(0).toUpperCase()}
                    </Avatar>
                    <div className="text-left">
                      <Text size="sm" fw={500} className="text-gray-900">
                        {currentWorkspace.name}
                      </Text>
                    </div>
                  </div>
                ) : (
                  "ワークスペースを選択"
                )}
              </Button>
            </Menu.Target>

            <Menu.Dropdown>
              <Menu.Label>ワークスペース</Menu.Label>
              {workspaces?.map((workspace: WorkspaceSummary) => (
                <Menu.Item
                  key={workspace.id}
                  onClick={() => handleWorkspaceChange(workspace.id)}
                  className={`${workspace.id === currentWorkspaceId ? "bg-blue-50" : ""}`}
                >
                  <div className="flex items-center space-x-2">
                    <Avatar size="sm" color="blue">
                      {workspace.name.charAt(0).toUpperCase()}
                    </Avatar>
                    <div>
                      <Text size="sm" fw={500}>
                        {workspace.name}
                      </Text>
                      {workspace.description && (
                        <Text size="xs" c="dimmed">
                          {workspace.description}
                        </Text>
                      )}
                    </div>
                  </div>
                </Menu.Item>
              ))}
            </Menu.Dropdown>
          </Menu>
        </div>

        {/* 中央に検索バー */}
        {isInWorkspace && (
          <div className="flex-1 max-w-2xl mx-auto">
            <form onSubmit={handleSearchSubmit}>
              <TextInput
                placeholder="メッセージ、チャンネル、ユーザーを検索"
                leftSection={<IconSearch size={16} />}
                value={searchQuery}
                onChange={(event) => setSearchQuery(event.currentTarget.value)}
                onFocus={handleSearchFocus}
                className="w-full"
              />
            </form>
          </div>
        )}

        <div className="flex items-center space-x-4 shrink-0">
          {/* ブックマークボタン */}
          {isInWorkspace && (
            <ActionIcon
              variant="subtle"
              size="lg"
              onClick={handleBookmarkClick}
              className="text-gray-700 hover:bg-gray-100"
              title="ブックマーク"
            >
              <IconBookmark size={20} />
            </ActionIcon>
          )}

          {/* ログアウトボタン */}
          <Button
            variant="subtle"
            onClick={handleLogout}
            className="text-gray-700 hover:bg-gray-100"
          >
            ログアウト
          </Button>
        </div>
      </div>
    </header>
  );
};
