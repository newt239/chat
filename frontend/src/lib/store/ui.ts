import { atom } from "jotai";
import { atomWithStorage } from "jotai/utils";

// メンバーパネルの開閉状態
export const isMemberPanelOpenAtom = atomWithStorage<boolean>(
  "ui-storage:isMemberPanelOpen",
  true
);

// メンバーパネルのトグル
export const toggleMemberPanelAtom = atom(null, (get, set) => {
  set(isMemberPanelOpenAtom, !get(isMemberPanelOpenAtom));
});

// メンバーパネルの開閉を設定
export const setMemberPanelOpenAtom = atom(
  null,
  (_get, set, open: boolean) => {
    set(isMemberPanelOpenAtom, open);
  }
);
