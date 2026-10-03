import { afterEach, expect, test, vi } from "vite-plus/test";

import { showNotification } from "#/features/notification/utils/notify";

afterEach(() => {
  vi.unstubAllGlobals();
});

const stubNotification = (permission: NotificationPermission) => {
  const created = vi.fn();
  // new で呼ばれるため class で差し替える
  class FakeNotification extends EventTarget {
    public static permission = permission;
    public constructor(title: string, options: NotificationOptions) {
      super();
      created(title, options);
    }
  }
  vi.stubGlobal("Notification", FakeNotification);
  return created;
};

test("許可があれば通知を出す", async () => {
  const created = stubNotification("granted");

  await showNotification({ body: "本文", onClick: () => {}, tag: "m1", title: "題" });

  expect(created).toHaveBeenCalledWith("題", { body: "本文", tag: "m1" });
});

test("許可がなければ通知を出さない", async () => {
  const created = stubNotification("denied");

  await showNotification({ body: "本文", onClick: () => {}, tag: "m1", title: "題" });

  expect(created).not.toHaveBeenCalled();
});
