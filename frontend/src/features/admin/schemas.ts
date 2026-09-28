import { z } from "zod";

export const adminTabValues = ["overview", "members", "permissions", "audit"] as const;

export const auditPeriodValues = ["day", "week", "month", "quarter", "all", "custom"] as const;

export const auditActionKeyValues = [
  "login",
  "loginFailed",
  "memberRoleChanged",
  "memberSuspended",
  "memberResumed",
  "channelCreated",
  "channelDeleted",
  "channelArchived",
  "channelUnarchived",
  "permissionChanged",
  "auditLogExported",
] as const;

// <input type="datetime-local"> の値（ブラウザのタイムゾーンでの日時）
const localDateTime = z.string().regex(/^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}$/u);

// 監査ログの絞り込みは tab=audit のときだけ使う。since / until は period=custom のときだけ見る
export const adminSearchSchema = z.object({
  action: z.enum(auditActionKeyValues).optional().catch(undefined),
  actor: z.string().optional().catch(undefined),
  period: z.enum(auditPeriodValues).optional().catch(undefined),
  since: localDateTime.optional().catch(undefined),
  tab: z.enum(adminTabValues).default("overview").catch("overview"),
  until: localDateTime.optional().catch(undefined),
});
