import { expect, test, vi } from "vite-plus/test";

import { navigateTo, registerRouter } from "#/lib/navigation";

test("router 登録前の navigateTo は何もしない", () => {
  expect(() => {
    navigateTo("/login");
  }).not.toThrow();
});

test("登録した router の navigate に遷移先を渡す", () => {
  const navigate = vi.fn(async () => {});
  registerRouter({ navigate });

  navigateTo("/login");

  expect(navigate).toHaveBeenCalledWith("/login");
});
