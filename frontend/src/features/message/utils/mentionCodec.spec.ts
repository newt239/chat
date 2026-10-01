import { describe, expect, test } from "vite-plus/test";

import { decodeMentions, encodeMentions } from "./mentionCodec";

const bob = "11111111-1111-4111-8111-111111111111";
const bobby = "22222222-2222-4222-8222-222222222222";
const general = "33333333-3333-4333-8333-333333333333";

describe("encodeMentions", () => {
  test("登録した名前と @channel を ID 記法にする", () => {
    const labels = new Map([
      ["@Bob", `<@${bob}>`],
      ["@Bob Smith", `<@${bobby}>`],
      ["#general", `<#${general}>`],
    ]);

    expect(encodeMentions("@Bob Smith と @Bob、@channel #general", labels)).toBe(
      `<@${bobby}> と <@${bob}>、<@channel> <#${general}>`,
    );
  });

  test("名前の続きに英数字があれば別の語として扱う", () => {
    const labels = new Map([["@Bob", `<@${bob}>`]]);

    expect(encodeMentions("@Bobby @channels", labels)).toBe("@Bobby @channels");
  });
});

describe("decodeMentions", () => {
  test("ID 記法を名前にし、戻せるよう対応を覚える", () => {
    const labels = new Map<string, string>();
    const text = decodeMentions(`<@${bob}> <@here> <#${general}>`, labels, (token) =>
      token.kind === "user" ? "@Bob" : token.kind === "broadcast" ? `@${token.id}` : null,
    );

    expect(text).toBe(`@Bob @here <#${general}>`);
    expect(encodeMentions(text, labels)).toBe(`<@${bob}> <@here> <#${general}>`);
  });
});
