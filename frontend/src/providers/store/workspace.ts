import { atomWithStorage } from "jotai/utils";

// 前回開いていたワークスペース。ログイン直後や /app を開いたときの移動先にする
export const lastWorkspaceIdAtom = atomWithStorage<string | null>(
  "last-workspace-id",
  null,
  undefined,
  { getOnInit: true },
);
