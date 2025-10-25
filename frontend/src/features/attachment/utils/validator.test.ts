import { describe, it, expect } from "vitest";

import { validateFile, formatFileSize } from "./validator";

describe("validateFile", () => {
  it("1GB 以下のファイルは valid", () => {
    const file = new File(["a".repeat(1000)], "test.txt", { type: "text/plain" });
    const result = validateFile(file);
    expect(result.valid).toBe(true);
  });

  it("1GB を超えるファイルは invalid", () => {
    const largeSize = 1024 * 1024 * 1024 + 1; // 1GB + 1 byte
    const file = new File([""], "large.txt", { type: "text/plain" });
    Object.defineProperty(file, "size", { value: largeSize });

    const result = validateFile(file);
    expect(result.valid).toBe(false);
    if (!result.valid) {
      expect(result.error).toContain("1GB");
    }
  });

  it("空のファイルは invalid", () => {
    const file = new File([], "empty.txt", { type: "text/plain" });
    const result = validateFile(file);
    expect(result.valid).toBe(false);
    if (!result.valid) {
      expect(result.error).toContain("空");
    }
  });
});

describe("formatFileSize", () => {
  it("0 バイトを正しくフォーマット", () => {
    expect(formatFileSize(0)).toBe("0 B");
  });

  it("バイト単位を正しくフォーマット", () => {
    expect(formatFileSize(500)).toBe("500.00 B");
  });

  it("KB 単位を正しくフォーマット", () => {
    expect(formatFileSize(1024)).toBe("1.00 KB");
    expect(formatFileSize(1536)).toBe("1.50 KB");
  });

  it("MB 単位を正しくフォーマット", () => {
    expect(formatFileSize(1024 * 1024)).toBe("1.00 MB");
    expect(formatFileSize(1.5 * 1024 * 1024)).toBe("1.50 MB");
  });

  it("GB 単位を正しくフォーマット", () => {
    expect(formatFileSize(1024 * 1024 * 1024)).toBe("1.00 GB");
  });
});
