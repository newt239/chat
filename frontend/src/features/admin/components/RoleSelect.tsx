import { useTranslation } from "react-i18next";

import { Select } from "#/components/ui/Select/Select";
import {
  assignableWorkspaceRoles,
  workspaceRoleKey,
} from "#/features/member/utils/workspaceRoleKeys";

import type { WorkspaceRole } from "#/gen/chat/v1/workspace_service_pb";

type RoleSelectProps = {
  ariaLabel: string;
  value: WorkspaceRole;
  onChange: (role: WorkspaceRole) => void;
  isDisabled: boolean;
};

export const RoleSelect = ({ ariaLabel, value, onChange, isDisabled }: RoleSelectProps) => {
  const { t } = useTranslation();
  return (
    <Select
      label={ariaLabel}
      isCompact
      options={assignableWorkspaceRoles.map(({ key }) => ({
        label: t(`member.role.${key}`),
        value: key,
      }))}
      value={workspaceRoleKey(value)}
      onChange={(key) => {
        const next = assignableWorkspaceRoles.find((option) => option.key === key);
        if (next !== undefined && next.role !== value) {
          onChange(next.role);
        }
      }}
      isDisabled={isDisabled}
    />
  );
};
