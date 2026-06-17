import * as Popover from "@radix-ui/react-popover";
import { CalendarDays, X } from "lucide-react";
import { useState } from "react";

import { Calendar } from "@/components/ui/calendar";
import { cn } from "@/lib/cn";

interface DatePickerProps {
  id?: string;
  /** Selected date as a `YYYY-MM-DD` string, or empty when unset. */
  value?: string;
  onChange(value: string): void;
  placeholder?: string;
  ariaDescribedby?: string;
  disabled?: boolean;
  className?: string;
}

function parseDate(value?: string): Date | undefined {
  if (!value) {
    return undefined;
  }
  const [y, m, d] = value.split("-").map(Number);
  if (!y || !m || !d) {
    return undefined;
  }
  return new Date(y, m - 1, d);
}

/** Formats a Date to a local `YYYY-MM-DD` string (no timezone shift). */
function formatValue(date: Date): string {
  const y = date.getFullYear();
  const m = String(date.getMonth() + 1).padStart(2, "0");
  const d = String(date.getDate()).padStart(2, "0");
  return `${y}-${m}-${d}`;
}

function formatLabel(date: Date): string {
  return date.toLocaleDateString(undefined, {
    year: "numeric",
    month: "long",
    day: "numeric",
  });
}

/** A button that opens a themed calendar popover. Works with YYYY-MM-DD strings. */
export function DatePicker({
  id,
  value,
  onChange,
  placeholder = "Pick a date",
  ariaDescribedby,
  disabled,
  className,
}: DatePickerProps) {
  const [open, setOpen] = useState(false);
  const selected = parseDate(value);

  return (
    <div className="flex items-center gap-1.5">
      <Popover.Root open={open} onOpenChange={setOpen}>
        <Popover.Trigger asChild>
          <button
            type="button"
            id={id}
            disabled={disabled}
            aria-describedby={ariaDescribedby}
            className={cn(
              "flex h-10 w-60 items-center gap-2 rounded-md border border-border bg-surface px-3 text-left text-sm transition-colors duration-150 hover:bg-muted focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 focus-visible:ring-offset-background disabled:cursor-not-allowed disabled:opacity-50 disabled:hover:bg-surface",
              !selected && "text-muted-foreground",
              className,
            )}
          >
            <CalendarDays className="size-4 shrink-0 text-muted-foreground" />
            <span className="truncate">
              {selected ? formatLabel(selected) : placeholder}
            </span>
          </button>
        </Popover.Trigger>
        {/* Not portaled: keeping the popover inside the Dialog's DOM subtree
            stops a day click from registering as an outside-click that would
            dismiss a surrounding modal. */}
        <Popover.Content
          side="top"
          align="start"
          sideOffset={8}
          avoidCollisions={false}
          className="z-50 rounded-lg border border-border bg-surface p-2 text-foreground shadow-md data-[state=open]:animate-[fade-in_140ms_ease-out]"
        >
          <Calendar
            mode="single"
            fixedWeeks
            selected={selected}
            defaultMonth={selected}
            autoFocus
            onSelect={(date) => {
              onChange(date ? formatValue(date) : "");
              if (date) {
                setOpen(false);
              }
            }}
          />
        </Popover.Content>
      </Popover.Root>
      {selected && !disabled ? (
        <button
          type="button"
          aria-label="Clear date"
          onClick={() => onChange("")}
          className="flex size-8 items-center justify-center rounded-md text-muted-foreground transition-colors duration-150 hover:bg-muted hover:text-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
        >
          <X className="size-4" />
        </button>
      ) : null}
    </div>
  );
}
