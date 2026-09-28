import { useTranslation } from "react-i18next";

import { permissionKeyByName, roleKeyByName } from "#/features/admin/utils/labels";
import { AuditAction } from "#/gen/chat/v1/admin_service_pb";

import type { AuditLog } from "#/gen/chat/v1/admin_service_pb";

type Metadata = AuditLog["metadata"];

const exportKindKeys: Readonly<Record<string, "auditLog" | "messages">> = {
  audit_log: "auditLog",
  messages: "messages",
};

// 監査ログの対象と詳細は内部名（ロール名や権限名）で届くので、表示用の文言に置き換える
export const useAuditLogFormatter = () => {
  const { t } = useTranslation();

  const roleName = (name: string | undefined) => {
    const key = name === undefined ? undefined : roleKeyByName[name];
    return key === undefined ? (name ?? "") : t(`member.role.${key}`);
  };

  const target = (log: AuditLog) => {
    if (log.targetType === "role") {
      return roleName(log.targetLabel);
    }
    const kind = log.metadata.kind === undefined ? undefined : exportKindKeys[log.metadata.kind];
    if (log.targetType === "data" && kind !== undefined) {
      return t(`admin.audit.detail.exportKinds.${kind}`);
    }
    return log.targetLabel || log.targetId;
  };

  const detailFormatters: Partial<Record<AuditAction, (metadata: Metadata) => string>> = {
    [AuditAction.MEMBER_ROLE_CHANGED]: ({ from, to }) => `${roleName(from)} → ${roleName(to)}`,
    [AuditAction.PERMISSION_CHANGED]: ({ permission, allowed }) => {
      const key = permission === undefined ? undefined : permissionKeyByName[permission];
      const name = key === undefined ? (permission ?? "") : t(`admin.permissions.names.${key}`);
      const state =
        allowed === "true" ? t("admin.audit.detail.allowed") : t("admin.audit.detail.denied");
      return `${name}: ${state}`;
    },
    [AuditAction.DATA_EXPORTED]: ({ count }) =>
      count === undefined ? "" : t("admin.audit.detail.exportCount", { count: Number(count) }),
    [AuditAction.CHANNEL_CREATED]: (metadata) =>
      metadata.private === "true" ? t("admin.audit.detail.private") : "",
  };
  const detail = (log: AuditLog) => detailFormatters[log.action]?.(log.metadata) ?? "";

  return { detail, target };
};
