import { render, screen } from "@testing-library/react";
import { Provider } from "jotai";
import { describe, it, expect } from "vitest";

import { ProfileSettingsPanel } from "./ProfileSettingsPanel";

import { MantineTestWrapper } from "@/test/MantineTestWrapper";


describe("ProfileSettingsPanel", () => {
  it("renders form fields", () => {
    render(
      <Provider>
        <MantineTestWrapper>
          <ProfileSettingsPanel />
        </MantineTestWrapper>
      </Provider>
    );
    expect(screen.getByLabelText("表示名")).toBeInTheDocument();
    expect(screen.getByLabelText("自己紹介")).toBeInTheDocument();
    expect(screen.getByLabelText("アイコンURL")).toBeInTheDocument();
  });
});


