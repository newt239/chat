import type { ReactNode } from "react";

import { Form } from "react-aria-components";

import { Button } from "#/components/ui/Button/Button";

type PasswordAuthFormProps = {
  onSubmit: () => void;
  error: Error | null;
  isPending: boolean;
  submitLabel: string;
  children: ReactNode;
};

export const PasswordAuthForm = ({
  onSubmit,
  error,
  isPending,
  submitLabel,
  children,
}: PasswordAuthFormProps) => (
  <Form
    className="flex flex-col gap-4"
    onSubmit={(event) => {
      event.preventDefault();
      onSubmit();
    }}
  >
    {children}
    {error !== null && <p className="m-0 text-caption text-danger">{error.message}</p>}
    <Button type="submit" isPending={isPending}>
      {submitLabel}
    </Button>
  </Form>
);
