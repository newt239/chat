import { render, screen } from "@testing-library/react";
import { Provider } from "jotai";
import { describe, it, expect } from "vitest";

import { NotificationPanel } from "./NotificationPanel";

describe("NotificationPanel", () => {
  it("通知がない場合、メッセージを表示する", () => {
    render(
      <Provider>
        <NotificationPanel />
      </Provider>
    );

    expect(screen.getByText("通知はありません")).toBeInTheDocument();
  });
});
