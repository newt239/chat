import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, test } from "vite-plus/test";

import { WebhookUrlPanel } from "./WebhookUrlPanel";

const url = "http://localhost:8080/webhooks/w1/token";

describe("WebhookUrlPanel", () => {
  test("URL と送信例を表示し、URL をコピーできる", async () => {
    const user = userEvent.setup();
    render(<WebhookUrlPanel url={url} />);
    expect(screen.getByText(url)).toBeInTheDocument();
    expect(screen.getByText(/curl -X POST/)).toHaveTextContent(url);

    await user.click(screen.getByRole("button", { name: "コピー" }));
    await waitFor(async () => {
      expect(await navigator.clipboard.readText()).toBe(url);
    });
  });
});
