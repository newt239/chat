import type * as jest from "@jest/expect";
import type { TestingLibraryMatchers } from "@testing-library/jest-dom/matchers";
import "vitest";

declare module "vitest" {
  type Assertion<T = unknown> = {} & jest.Matchers<void, T> & TestingLibraryMatchers<T, void>
}
