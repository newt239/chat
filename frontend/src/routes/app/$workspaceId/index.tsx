import { createFileRoute } from "@tanstack/react-router";

const WorkspaceIndexComponent = () => {
  return (
    <div className="flex h-full items-center justify-center">
      <div className="text-center">
        <h1 className="text-2xl font-bold text-gray-900 mb-4">チャンネルを選択してください</h1>
        <p className="text-gray-600">
          左側のサイドバーからチャンネルを選択してください。
        </p>
      </div>
    </div>
  );
};

export const Route = createFileRoute("/app/$workspaceId/")({
  component: WorkspaceIndexComponent,
});
