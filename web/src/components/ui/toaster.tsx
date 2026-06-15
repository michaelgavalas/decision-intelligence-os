import { Toaster as SonnerToaster, toast } from "sonner";

import { useTheme } from "@/hooks/useTheme";

/** Themed application toaster. Mount once near the app root. */
export function Toaster() {
  const { resolvedTheme } = useTheme();

  return (
    <SonnerToaster
      theme={resolvedTheme}
      position="bottom-right"
      richColors
      closeButton
      toastOptions={{
        classNames: {
          toast: "font-sans",
        },
      }}
    />
  );
}

/** Re-export of sonner's imperative toast API. */
export const useToast = () => toast;
export { toast };
