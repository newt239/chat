import { render, screen } from "@testing-library/react";
import { describe, expect, test } from "vite-plus/test";

import { AppUrlPanel } from "./AppUrlPanel";

describe("AppUrlPanel", () => {
  test("URL と送信例を出す", () => {
    render(<AppUrlPanel url="https://api.example.com/webhooks/a1/t" />);

    expect(
      screen.getByText("https://api.example.com/webhooks/a1/t", { selector: "code" }),
    ).toBeInTheDocument();
    expect(screen.getByText(/curl -X POST/)).toHaveTextContent("channel_id");
  });
});
