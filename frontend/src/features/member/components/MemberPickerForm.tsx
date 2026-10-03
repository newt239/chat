import { useState } from "react";

import { Form } from "react-aria-components";

import { Button } from "#/components/ui/Button/Button";
import { ComboBox } from "#/components/ui/ComboBox/ComboBox";
import { useDisplayName } from "#/features/member/hooks/useDisplayName";
import { useMembers } from "#/features/member/hooks/useMembers";

type MemberPickerFormProps = {
  workspaceId: string;
  // 候補から除くメンバー
  memberIds: ReadonlySet<string>;
  label: string;
  placeholder: string;
  submitLabel: string;
  isPending: boolean;
  onSubmit: (userId: string) => void;
};

// ワークスペースのメンバーから 1 人選んで追加するフォーム
export const MemberPickerForm = ({
  workspaceId,
  memberIds,
  label,
  placeholder,
  submitLabel,
  isPending,
  onSubmit,
}: MemberPickerFormProps) => {
  const { data: workspaceMembers } = useMembers(workspaceId);
  const displayName = useDisplayName();
  const [selectedUserId, setSelectedUserId] = useState<string | null>(null);
  const options = (workspaceMembers ?? [])
    .filter((member) => !memberIds.has(member.userId))
    .map((member) => ({
      label: displayName(member.userId, member.displayName),
      value: member.userId,
    }));

  return (
    <Form
      className="flex items-end gap-2"
      onSubmit={(event) => {
        event.preventDefault();
        if (selectedUserId !== null) {
          onSubmit(selectedUserId);
          setSelectedUserId(null);
        }
      }}
    >
      <ComboBox
        label={label}
        placeholder={placeholder}
        options={options}
        value={selectedUserId}
        onChange={setSelectedUserId}
        className="flex-1"
      />
      <Button
        type="submit"
        variant="secondary"
        isDisabled={selectedUserId === null}
        isPending={isPending}
      >
        {submitLabel}
      </Button>
    </Form>
  );
};
