import { render, screen } from "@testing-library/react";
import { expect, test } from "vite-plus/test";

import { ChannelName } from "./ChannelName";

test("階層のパスは親を区切って表示し、公開範囲をアイコンで示す", () => {
  const { container } = render(<ChannelName name="dev/frontend" isPrivate />);

  expect(container).toHaveTextContent("dev / frontend");
  expect(screen.getByRole("img", { name: "非公開チャンネル" })).toBeInTheDocument();
});
