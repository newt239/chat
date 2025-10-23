import { WorkspaceList } from "@/features/workspace/components/WorkspaceList";

export const WorkspaceSelection = () => {
  console.log("WorkspaceSelection - コンポーネントがレンダリングされました");

  return (
    <div className="flex h-full items-center justify-center">
      <div className="text-center">
        <h1 className="text-2xl font-bold text-gray-900 mb-4">ワークスペースを選択してください</h1>
        <p className="text-gray-600 mb-8">
          参加しているワークスペースから選択するか、新しいワークスペースを作成してください。
        </p>
        <WorkspaceList />
      </div>
    </div>
  );
};
