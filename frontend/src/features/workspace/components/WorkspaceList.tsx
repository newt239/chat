import { useEffect, useState } from "react";

import { Card, Text, Button, Group, Stack, Loader } from "@mantine/core";
import { useAtomValue, useSetAtom } from "jotai";

import { useWorkspaces } from "../hooks/useWorkspace";

import { CreateWorkspaceModal } from "./CreateWorkspaceModal";

import { router } from "@/lib/router";
import { currentWorkspaceIdAtom, setCurrentWorkspaceAtom } from "@/providers/store/workspace";

export const WorkspaceList = () => {
  const { data, isLoading, error } = useWorkspaces();
  const [isModalOpen, setIsModalOpen] = useState(false);
  const currentWorkspaceId = useAtomValue(currentWorkspaceIdAtom);
  const setCurrentWorkspace = useSetAtom(setCurrentWorkspaceAtom);

  useEffect(() => {
    if (data && data.length > 0 && currentWorkspaceId === null) {
      const firstWorkspace = data[0];
      if (firstWorkspace) {
        setCurrentWorkspace(firstWorkspace.id);
      }
    }
  }, [data, currentWorkspaceId, setCurrentWorkspace]);

  if (isLoading) {
    return (
      <div className="flex justify-center items-center h-64">
        <Loader />
      </div>
    );
  }

  if (error) {
    return (
      <Text c="red" className="text-center">
        ワークスペースの読み込みに失敗しました
      </Text>
    );
  }

  return (
    <>
      <Stack gap="md">
        <Group justify="space-between">
          <Text size="lg" fw={500}>
            あなたのワークスペース
          </Text>
          <Button onClick={() => setIsModalOpen(true)}>新規作成</Button>
        </Group>

        {data && Array.isArray(data) && data.length > 0 && (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
            {data.map((workspace) => {
              const isSelected = workspace.id === currentWorkspaceId;
              return (
                <Card
                  key={workspace.id}
                  shadow="sm"
                  padding="lg"
                  radius="md"
                  withBorder
                  className={isSelected ? "border-blue-500" : undefined}
                >
                  <Text fw={500} size="lg" className="mb-2">
                    {workspace.name}
                  </Text>
                  {workspace.description && (
                    <Text size="sm" c="dimmed" className="mb-4">
                      {workspace.description}
                    </Text>
                  )}
                  <Button
                    variant={isSelected ? "filled" : "light"}
                    fullWidth
                    onClick={() => {
                      setCurrentWorkspace(workspace.id);
                      router.navigate({
                        to: "/app/$workspaceId",
                        params: { workspaceId: workspace.id },
                      });
                    }}
                  >
                    {isSelected ? "選択中" : "開く"}
                  </Button>
                </Card>
              );
            })}
          </div>
        )}
      </Stack>

      <CreateWorkspaceModal opened={isModalOpen} onClose={() => setIsModalOpen(false)} />
    </>
  );
};
