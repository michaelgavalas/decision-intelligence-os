import * as SliderPrimitive from "@radix-ui/react-slider";
import { forwardRef } from "react";

import { cn } from "@/lib/cn";

/**
 * A draggable range slider (Radix). Pass `value`/`onValueChange` as arrays, e.g.
 * `value={[70]} onValueChange={([v]) => ...}`. The thumb is keyboard-operable.
 */
export const Slider = forwardRef<
  React.ComponentRef<typeof SliderPrimitive.Root>,
  React.ComponentPropsWithoutRef<typeof SliderPrimitive.Root> & {
    thumbLabel?: string;
  }
>(({ className, thumbLabel, ...props }, ref) => (
  <SliderPrimitive.Root
    ref={ref}
    className={cn(
      "relative flex w-full touch-none select-none items-center py-1.5",
      className,
    )}
    {...props}
  >
    <SliderPrimitive.Track className="relative h-2 w-full grow overflow-hidden rounded-full bg-muted">
      <SliderPrimitive.Range className="absolute h-full bg-primary" />
    </SliderPrimitive.Track>
    <SliderPrimitive.Thumb
      aria-label={thumbLabel}
      className="block size-5 cursor-grab rounded-full border-2 border-primary bg-surface shadow-sm transition-[box-shadow,transform] duration-150 hover:scale-105 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 focus-visible:ring-offset-background active:cursor-grabbing"
    />
  </SliderPrimitive.Root>
));
Slider.displayName = "Slider";
