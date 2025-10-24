import { useState } from "react";

import { Menu, Button, Text, Avatar } from "@mantine/core";

import { useWorkspaces } from "@/features/workspace/hooks/useWorkspace";
import { useAuthStore } from "@/lib/store/auth";
import { useWorkspaceStore } from "@/lib/store/workspace";

interface Workspace {
  id: string;
  name: string;
  description?: string | null;
}

export const Header = () => {
  const { data: workspaces, isLoading } = useWorkspaces();
  const currentWorkspaceId = useWorkspaceStore((state) => state.currentWorkspaceId);
  const setCurrentWorkspace = useWorkspaceStore((state) => state.setCurrentWorkspace);
  const clearAuth = useAuthStore((state) => state.clearAuth);
  const [isWorkspaceMenuOpen, setIsWorkspaceMenuOpen] = useState(false);

  const currentWorkspace = workspaces?.find((w: Workspace) => w.id === currentWorkspaceId);

  const handleWorkspaceChange = (workspaceId: string) => {
    setCurrentWorkspace(workspaceId);
    setIsWorkspaceMenuOpen(false);
  };

  const handleLogout = () => {
    clearAuth();
  };

  return (
    <header className="bg-white border-b border-gray-200 px-4 py-3">
      <div className="flex items-center justify-between">
        <div className="flex items-center space-x-4">
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
              {workspaces?.map((workspace: Workspace) => (
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

        <div className="flex items-center space-x-4">
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
