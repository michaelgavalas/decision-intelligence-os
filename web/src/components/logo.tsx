import { cn } from "@/lib/cn";

interface LogoProps {
  className?: string;
}

/**
 * Typographic wordmark for Decision Intelligence OS. Set in the display
 * typeface with a primary-colored "OS" lockup so the product name reads as a
 * brand rather than a generic app title.
 */
export function Logo({ className }: LogoProps) {
  return (
    <span
      className={cn(
        "font-display text-[15px] font-semibold leading-none tracking-tight whitespace-nowrap text-foreground select-none",
        className,
      )}
    >
      Decision Intelligence<span className="text-primary">&nbsp;OS</span>
    </span>
  );
}
