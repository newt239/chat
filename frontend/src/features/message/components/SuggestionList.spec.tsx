import { fireEvent, render, screen } from "@testing-library/react";
import { describe, expect, test, vi } from "vite-plus/test";

import { SuggestionList } from "./SuggestionList";

import type { SuggestionItem } from "../utils/suggestion";

const items: SuggestionItem[] = [
  {
    avatarUrl: undefined,
    id: "u1",
    kind: "user",
    label: "Alice Johnson",
    token: "<@u1>",
    value: "@Alice Johnson",
  },
  {
    avatarUrl: undefined,
    id: "g1",
    kind: "group",
    label: "developers",
    token: "<@&g1>",
    value: "@developers",
  },
  {
    avatarUrl: undefined,
    id: "c1",
    kind: "channel",
    label: "dev/frontend",
    token: "<#c1>",
    value: "#dev/frontend",
  },
];

describe("SuggestionList", () => {
  test("選択中の候補を示し、押した候補を選ぶ", () => {
    const onSelect = vi.fn<(item: SuggestionItem) => void>();
    render(<SuggestionList id="list" items={items} activeIndex={1} onSelect={onSelect} />);

    expect(screen.getByRole("listbox", { name: "候補" })).toBeInTheDocument();
    expect(screen.getByRole("option", { name: /developers/ })).toHaveAttribute(
      "aria-selected",
      "true",
    );
    expect(screen.getByRole("option", { name: /Alice Johnson/ })).toBeInTheDocument();

    fireEvent.pointerDown(screen.getByRole("option", { name: /dev\/frontend/ }));
    expect(onSelect).toHaveBeenCalledWith(items[2]);
  });
});
