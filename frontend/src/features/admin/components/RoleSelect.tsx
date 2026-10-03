import { IconCheck, IconChevronDown } from "@tabler/icons-react";
import { Button, ListBox, ListBoxItem, Popover, Select, SelectValue } from "react-aria-components";
import { useTranslation } from "react-i18next";

import { cn, focusRing, overlayStyles } from "#/components/ui/styles/styles";
import { workspaceRoles } from "#/features/member/utils/workspaceRoleKeys";
import { WorkspaceRole } from "#/gen/chat/v1/workspace_service_pb";

const roles = workspaceRoles.filter((option) => option.role !== WorkspaceRole.OWNER);

type RoleSelectProps = {
  ariaLabel: string;
  value: WorkspaceRole;
  onChange: (role: WorkspaceRole) => void;
  isDisabled: boolean;
};

// 表の中に置くため見出しを持たない小さなセレクト
export const RoleSelect = ({ ariaLabel, value, onChange, isDisabled }: RoleSelectProps) => {
  const { t } = useTranslation();
  return (
    <Select
      aria-label={ariaLabel}
      value={roles.find((option) => option.role === value)?.key ?? null}
      onChange={(key) => {
        const next = roles.find((option) => option.key === key);
        if (next !== undefined && next.role !== value) {
          onChange(next.role);
        }
      }}
      isDisabled={isDisabled}
      className="font-sans"
    >
      <Button
        className={cn(
          "flex h-7 min-w-24 cursor-pointer items-center gap-1.5 rounded-md border border-border-strong bg-surface px-2 text-left text-label font-normal text-text data-disabled:cursor-default data-disabled:bg-sunken data-disabled:text-subtle",
          focusRing,
        )}
      >
        <SelectValue className="flex-1 truncate" />
        <IconChevronDown aria-hidden className="size-3.5 shrink-0 text-muted" />
      </Button>
      <Popover offset={4} className={cn(overlayStyles.popover, "min-w-(--trigger-width) p-1")}>
        <ListBox items={roles} className="outline-none">
          {(option) => (
            <ListBoxItem
              id={option.key}
              textValue={t(`member.role.${option.key}`)}
              className={overlayStyles.listItem}
            >
              {({ isSelected }) => (
                <>
                  <span className="flex-1">{t(`member.role.${option.key}`)}</span>
                  {isSelected && <IconCheck aria-hidden />}
                </>
              )}
            </ListBoxItem>
          )}
        </ListBox>
      </Popover>
    </Select>
  );
};
