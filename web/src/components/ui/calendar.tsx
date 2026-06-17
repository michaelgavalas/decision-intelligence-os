import { DayPicker, type DayPickerProps } from "react-day-picker";
import "react-day-picker/style.css";

import { cn } from "@/lib/cn";

/**
 * Date picker calendar (react-day-picker) themed with the app's design tokens.
 * Visual overrides live under the `.rdp-dios` scope in globals.css.
 */
export function Calendar({ className, ...props }: DayPickerProps) {
  return (
    <DayPicker showOutsideDays className={cn("rdp-dios", className)} {...props} />
  );
}
