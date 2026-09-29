import { create } from "@bufbuild/protobuf";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, test, vi } from "vite-plus/test";

import { MessageLocationSchema } from "#/gen/chat/v1/message_pb";

import { PendingLocation } from "./PendingLocation";

describe("PendingLocation", () => {
  test("送信前の位置情報を表示し、外せる", async () => {
    const onRemove = vi.fn<() => void>();
    render(
      <PendingLocation
        location={create(MessageLocationSchema, { label: "駅", latitude: 35, longitude: 139 })}
        onRemove={onRemove}
      />,
    );

    expect(screen.getByText("駅")).toBeInTheDocument();
    expect(screen.getByText("35.00000, 139.00000")).toBeInTheDocument();
    await userEvent.click(screen.getByRole("button", { name: "位置情報を外す" }));
    expect(onRemove).toHaveBeenCalled();
  });
});
