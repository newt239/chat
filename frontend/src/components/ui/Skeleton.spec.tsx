import { render } from "@testing-library/react";
import { describe, expect, test } from "vite-plus/test";

import { Skeleton } from "./Skeleton";

describe("Skeleton", () => {
  test("読み上げ対象から外し、指定した大きさと形で表示する", () => {
    const { container } = render(<Skeleton className="size-8 rounded-full" />);

    const skeleton = container.firstElementChild;
    expect(skeleton).toHaveAttribute("aria-hidden", "true");
    expect(skeleton).toHaveClass("size-8", "rounded-full", "animate-pulse");
    expect(skeleton).not.toHaveClass("rounded-md");
  });
});
