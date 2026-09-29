import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, test } from "vite-plus/test";

import { renderWithProviders } from "#/test/renderWithProviders";

import { jumpPresets } from "../utils/dateJump";
import { DateDivider } from "./DateDivider";

const presets = jumpPresets(new Date());

describe("DateDivider", () => {
  test("今日・昨日は言葉で、それ以外は曜日付きの日付で表示する", async () => {
    await renderWithProviders(
      <>
        <DateDivider dateKey={presets.today} />
        <DateDivider dateKey="2026-09-01" />
      </>,
      "/app/ws1/ch1",
      () => {},
    );

    expect(await screen.findByRole("button", { name: /^今日/ })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /^2026年9月1日\(火\)/ })).toBeInTheDocument();
  });

  test("メニューで選んだ日へ ?date= で移動し、?message= は外す", async () => {
    const { router } = await renderWithProviders(
      <DateDivider dateKey={presets.today} />,
      "/app/ws1/ch1?message=m1",
      () => {},
    );

    await userEvent.click(await screen.findByRole("button", { name: /^今日/ }));
    await userEvent.click(await screen.findByRole("menuitem", { name: "昨日" }));

    await waitFor(() => {
      expect(router.state.location.search).toEqual({ date: presets.yesterday });
    });
  });

  test("日付を指定するとカレンダーを開く", async () => {
    await renderWithProviders(<DateDivider dateKey={presets.today} />, "/app/ws1/ch1", () => {});

    await userEvent.click(await screen.findByRole("button", { name: /^今日/ }));
    await userEvent.click(await screen.findByRole("menuitem", { name: "日付を指定" }));

    expect(await screen.findByRole("application", { name: /日付を選ぶ/ })).toBeInTheDocument();
  });
});
