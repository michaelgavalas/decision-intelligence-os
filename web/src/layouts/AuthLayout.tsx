import { Outlet } from "react-router-dom";

import { Logo } from "@/components/logo";

export function AuthLayout() {
  return (
    <div className="flex min-h-dvh flex-col items-center justify-center bg-background px-4 py-12">
      <div className="mb-8">
        <Logo className="text-2xl" />
      </div>
      <div className="w-full max-w-md">
        <Outlet />
      </div>
    </div>
  );
}
