import { useState } from "react";

import { Modal, TextInput, Textarea, Button, Group } from "@mantine/core";
import { useForm } from "@mantine/form";

import { useUserGroups } from "../hooks/useUserGroups";

import type { CreateUserGroupInput } from "../types";

type CreateUserGroupModalProps = {
  opened: boolean;
  onClose: () => void;
  workspaceId: string;
  onSuccess: () => void;
}

export const CreateUserGroupModal = ({
  opened,
  onClose,
  workspaceId,
  onSuccess,
}: CreateUserGroupModalProps) => {
  const { createGroup } = useUserGroups(workspaceId);
  const [isLoading, setIsLoading] = useState(false);

  const form = useForm<CreateUserGroupInput>({
    initialValues: {
      workspaceId,
      name: "",
      description: "",
    },
    validate: {
      name: (value) => {
        if (!value.trim()) return "Group name is required";
        if (value.length < 2) return "Group name must be at least 2 characters";
        if (value.length > 50) return "Group name must be less than 50 characters";
        if (!/^[a-zA-Z0-9_-]+$/.test(value)) {
          return "Group name can only contain letters, numbers, hyphens, and underscores";
        }
        return null;
      },
      description: (value) => {
        if (value && value.length > 200) {
          return "Description must be less than 200 characters";
        }
        return null;
      },
    },
  });

  const handleSubmit = async (values: CreateUserGroupInput) => {
    setIsLoading(true);
    try {
      await createGroup(values);
      form.reset();
      onSuccess();
    } catch (error) {
      console.error("Failed to create group:", error);
    } finally {
      setIsLoading(false);
    }
  };

  const handleClose = () => {
    form.reset();
    onClose();
  };

  return (
    <Modal opened={opened} onClose={handleClose} title="Create User Group" size="md">
      <form onSubmit={form.onSubmit(handleSubmit)}>
        <div className="space-y-4">
          <TextInput
            label="Group Name"
            placeholder="e.g., developers, marketing"
            description="This will be used for @mentions (e.g., @developers)"
            required
            {...form.getInputProps("name")}
          />

          <Textarea
            label="Description"
            placeholder="Optional description of the group"
            rows={3}
            {...form.getInputProps("description")}
          />

          <Group justify="flex-end" mt="md">
            <Button variant="subtle" onClick={handleClose}>
              Cancel
            </Button>
            <Button type="submit" loading={isLoading}>
              Create Group
            </Button>
          </Group>
        </div>
      </form>
    </Modal>
  );
};
