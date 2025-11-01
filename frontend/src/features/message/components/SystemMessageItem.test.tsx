import { render, screen } from "@testing-library/react";

import { SystemMessageItem } from "./SystemMessageItem";

describe("SystemMessageItem", () => {
  it("renders member_joined text", () => {
    render(
      <SystemMessageItem
        message={{
          id: "1",
          channelId: "c1",
          kind: "member_joined",
          payload: { userId: "u1" },
          createdAt: new Date().toISOString(),
        }}
      />
    );
    expect(screen.getByText(/参加しました/)).toBeTruthy();
  });
});
