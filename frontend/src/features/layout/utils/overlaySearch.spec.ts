import { describe, expect, test } from "vite-plus/test";

import { closeDialog, closePanel, openDialog, openPanel } from "./overlaySearch";

describe("overlaySearch", () => {
  test("パネルを開くと他のパネルとダイアログを閉じ、ルートの search は残す", () => {
    expect(
      openPanel({ profile: "u1" })({ dialog: "create-dm", group: "g1", message: "m1" }),
    ).toEqual({ ...closeDialog({}), ...closePanel({}), message: "m1", profile: "u1" });
  });

  test("ダイアログを開いても右パネルは残す", () => {
    expect(openDialog({ dialog: "edit-group" })({ group: "g1", settings: "theme" })).toEqual(
      expect.objectContaining({ dialog: "edit-group", group: "g1", settings: undefined }),
    );
  });

  test("閉じるとそのキーを undefined にする", () => {
    expect(closeDialog({ emoji: "👍", profile: "u1", reactions: "m1" })).toEqual(
      expect.objectContaining({ emoji: undefined, profile: "u1", reactions: undefined }),
    );
    expect(closePanel({ panel: "pins", q: "x" })).toEqual(
      expect.objectContaining({ panel: undefined, q: "x" }),
    );
  });
});
