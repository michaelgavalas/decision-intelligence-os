import { Skeleton } from "@/components/ui/skeleton";
import { cn } from "@/lib/cn";

export function ChartLoading({ height }: { height: number }) {
  return <Skeleton style={{ height }} className="w-full" />;
}

export function ChartEmpty({
  height,
  message = "No data to display yet.",
}: {
  height: number;
  message?: string;
}) {
  return (
    <div
      style={{ height }}
      className={cn(
        "flex w-full items-center justify-center rounded-md border border-dashed border-border text-sm text-muted-foreground",
      )}
    >
      {message}
    </div>
  );
}
