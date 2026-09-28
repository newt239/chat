import { act, fireEvent, render, screen } from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, test, vi } from "vite-plus/test";

import { useLongPress } from "./useLongPress";

const Target = ({ onLongPress, isEnabled }: { onLongPress: () => void; isEnabled: boolean }) => {
  const { isPressed, longPressProps } = useLongPress(onLongPress, isEnabled);
  return (
    <div {...longPressProps} data-pressed={isPressed}>
      target
    </div>
  );
};

describe("useLongPress", () => {
  beforeEach(() => {
    vi.useFakeTimers();
  });
  afterEach(() => {
    vi.useRealTimers();
  });

  test("押し続けると呼ばれる", () => {
    const onLongPress = vi.fn<() => void>();
    render(<Target onLongPress={onLongPress} isEnabled />);

    fireEvent.pointerDown(screen.getByText("target"), { button: 0, clientX: 0, clientY: 0 });
    expect(screen.getByText("target")).toHaveAttribute("data-pressed", "true");
    act(() => {
      vi.advanceTimersByTime(500);
    });

    expect(onLongPress).toHaveBeenCalledOnce();
  });

  test("指を動かす（スクロールする）と取り消す", () => {
    const onLongPress = vi.fn<() => void>();
    render(<Target onLongPress={onLongPress} isEnabled />);

    const target = screen.getByText("target");
    fireEvent.pointerDown(target, { button: 0, clientX: 0, clientY: 0 });
    fireEvent.pointerMove(target, { clientX: 0, clientY: 20 });
    act(() => {
      vi.advanceTimersByTime(500);
    });

    expect(onLongPress).not.toHaveBeenCalled();
  });

  test("無効なときは反応しない", () => {
    const onLongPress = vi.fn<() => void>();
    render(<Target onLongPress={onLongPress} isEnabled={false} />);

    fireEvent.contextMenu(screen.getByText("target"));

    expect(onLongPress).not.toHaveBeenCalled();
  });
});
