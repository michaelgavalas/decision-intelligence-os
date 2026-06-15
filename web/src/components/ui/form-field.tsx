import { useId } from "react";

import { cn } from "@/lib/cn";

interface FormFieldProps {
  label: string;
  /** Receives the id and aria props to spread onto the control. */
  children: (props: {
    id: string;
    "aria-describedby"?: string;
    "aria-invalid"?: boolean;
    "aria-required"?: boolean;
  }) => React.ReactNode;
  helperText?: string;
  error?: string;
  required?: boolean;
  className?: string;
}

export function FormField({
  label,
  children,
  helperText,
  error,
  required,
  className,
}: FormFieldProps) {
  const id = useId();
  const helperId = `${id}-helper`;
  const errorId = `${id}-error`;

  const describedBy =
    [error ? errorId : null, helperText ? helperId : null]
      .filter(Boolean)
      .join(" ") || undefined;

  return (
    <div className={cn("flex flex-col gap-1.5", className)}>
      <label htmlFor={id} className="text-sm font-medium text-foreground">
        {label}
        {required ? (
          <span className="ml-0.5 text-destructive" aria-hidden="true">
            *
          </span>
        ) : null}
      </label>

      {children({
        id,
        "aria-describedby": describedBy,
        "aria-invalid": error ? true : undefined,
        "aria-required": required || undefined,
      })}

      {helperText && !error ? (
        <p id={helperId} className="text-xs text-muted-foreground">
          {helperText}
        </p>
      ) : null}

      {error ? (
        <p id={errorId} className="text-xs text-destructive">
          {error}
        </p>
      ) : null}
    </div>
  );
}
