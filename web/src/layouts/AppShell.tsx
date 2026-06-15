import { Suspense, useEffect, useState } from "react";
import { NavLink, Outlet, useLocation, useNavigate } from "react-router-dom";
import {
  BarChart3,
  Building2,
  Check,
  ListChecks,
  LogOut,
  PanelLeft,
  Users,
} from "lucide-react";
import type { LucideIcon } from "lucide-react";

import { Logo } from "@/components/logo";
import { ThemeToggle } from "@/components/theme-toggle";
import { Button } from "@/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { Spinner } from "@/components/ui/spinner";
import { useAuth } from "@/features/auth/use-auth";
import { TeamSwitcher } from "@/features/teams/TeamSwitcher";
import { useCurrentTeam } from "@/features/teams/use-current-team";
import { cn } from "@/lib/cn";

interface NavItem {
  to: string;
  label: string;
  icon: LucideIcon;
}

const NAV_ITEMS: NavItem[] = [
  { to: "/decisions", label: "Decisions", icon: ListChecks },
  { to: "/analytics", label: "Analytics", icon: BarChart3 },
  { to: "/teams", label: "Teams", icon: Users },
];

/** Derives up-to-two-letter initials from a display name (e.g. "Ada Lovelace" -> "AL"). */
function initials(name?: string): string {
  if (!name) {
    return "?";
  }
  const parts = name.trim().split(/\s+/).filter(Boolean);
  if (parts.length === 0) {
    return "?";
  }
  const first = parts[0][0] ?? "";
  const last = parts.length > 1 ? (parts[parts.length - 1][0] ?? "") : "";
  return (first + last).toUpperCase();
}

function NavItems({ onNavigate }: { onNavigate?: () => void }) {
  return (
    <nav className="flex flex-col gap-1" aria-label="Primary">
      {NAV_ITEMS.map(({ to, label, icon: Icon }) => (
        <NavLink
          key={to}
          to={to}
          onClick={onNavigate}
          className={({ isActive }) =>
            cn(
              "flex items-center gap-3 rounded-md px-3 py-2 text-sm font-medium transition-colors duration-150",
              isActive
                ? "bg-primary/10 text-primary"
                : "text-muted-foreground hover:bg-muted hover:text-foreground",
            )
          }
        >
          <Icon className="size-[18px]" />
          {label}
        </NavLink>
      ))}
    </nav>
  );
}

function PageFallback() {
  return (
    <div className="flex min-h-64 items-center justify-center">
      <Spinner className="size-8 text-muted-foreground" />
    </div>
  );
}

export function AppShell() {
  const [sidebarOpen, setSidebarOpen] = useState(
    () => typeof window === "undefined" || window.innerWidth >= 768,
  );
  const { user, logout } = useAuth();
  const { teams, currentTeamId, setCurrentTeamId } = useCurrentTeam();
  const navigate = useNavigate();
  const location = useLocation();

  // Close the mobile drawer on Escape for keyboard users.
  useEffect(() => {
    if (!sidebarOpen) {
      return;
    }
    function onKeyDown(event: KeyboardEvent) {
      if (event.key === "Escape" && window.innerWidth < 768) {
        setSidebarOpen(false);
      }
    }
    document.addEventListener("keydown", onKeyDown);
    return () => document.removeEventListener("keydown", onKeyDown);
  }, [sidebarOpen]);

  const handleLogout = async () => {
    await logout();
    navigate("/login", { replace: true });
  };

  // On mobile the sidebar is an overlay, so a nav tap should dismiss it.
  const closeOnMobile = () => {
    if (window.innerWidth < 768) {
      setSidebarOpen(false);
    }
  };

  return (
    <div className="flex h-dvh flex-col bg-background">
      <a
        href="#main"
        className="sr-only z-50 rounded-md bg-primary px-4 py-2 text-sm font-semibold text-primary-foreground focus:not-sr-only focus:absolute focus:left-4 focus:top-4 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 focus-visible:ring-offset-background"
      >
        Skip to main content
      </a>

      {/* Full-width top bar */}
      <header className="flex h-14 shrink-0 items-center gap-3 border-b border-border bg-surface px-3 sm:px-4">
        <Button
          variant="ghost"
          size="iconSm"
          aria-label={sidebarOpen ? "Hide sidebar" : "Show sidebar"}
          aria-expanded={sidebarOpen}
          onClick={() => setSidebarOpen((open) => !open)}
        >
          <PanelLeft className="size-4" />
        </Button>
        <Logo />

        <div className="ml-auto flex items-center gap-1.5">
          <div className="hidden md:block">
            <TeamSwitcher />
          </div>
          <ThemeToggle />
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <button
                type="button"
                aria-label="Account menu"
                className="flex size-8 items-center justify-center rounded-full bg-primary/10 text-xs font-semibold text-primary ring-1 ring-inset ring-primary/20 transition-colors duration-150 hover:bg-primary/15 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
              >
                {initials(user?.name)}
              </button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end">
              <DropdownMenuLabel>
                <span className="block text-sm font-semibold text-foreground">
                  {user?.name ?? "Account"}
                </span>
                {user?.email ? (
                  <span className="block truncate font-normal text-muted-foreground">
                    {user.email}
                  </span>
                ) : null}
              </DropdownMenuLabel>

              {/* Team switcher - surfaced here on mobile, where the top-bar
                  switcher is hidden. */}
              {teams.length > 0 ? (
                <div className="md:hidden">
                  <DropdownMenuSeparator />
                  <DropdownMenuLabel>Team</DropdownMenuLabel>
                  {teams.map((team) => (
                    <DropdownMenuItem
                      key={team.id}
                      onSelect={() => setCurrentTeamId(team.id)}
                    >
                      <Building2 className="size-4 text-muted-foreground" />
                      <span className="truncate">{team.name}</span>
                      {team.id === currentTeamId ? (
                        <Check className="ml-auto size-4 text-primary" />
                      ) : null}
                    </DropdownMenuItem>
                  ))}
                </div>
              ) : null}

              <DropdownMenuSeparator />
              <DropdownMenuItem
                destructive
                onSelect={() => {
                  void handleLogout();
                }}
              >
                <LogOut className="size-4" />
                Log out
              </DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>
        </div>
      </header>

      {/* Sidebar + content row */}
      <div className="flex min-h-0 flex-1">
        {/* Desktop sidebar (in-flow; smoothly collapses to zero width) */}
        <aside
          className={cn(
            "hidden shrink-0 overflow-hidden border-r border-border bg-surface transition-[width] duration-200 ease-out md:block",
            sidebarOpen ? "w-64" : "w-0 border-r-0",
          )}
          aria-hidden={!sidebarOpen}
        >
          <div className="h-full w-64 overflow-y-auto p-3">
            <NavItems />
          </div>
        </aside>

        {/* Mobile sidebar (overlay drawer below the top bar) */}
        {sidebarOpen ? (
          <div className="fixed inset-x-0 bottom-0 top-14 z-40 md:hidden">
            <button
              type="button"
              aria-label="Close sidebar"
              className="absolute inset-0 bg-black/50 animate-fade-in"
              onClick={() => setSidebarOpen(false)}
            />
            <aside className="absolute left-0 top-0 h-full w-64 animate-drawer-in overflow-y-auto border-r border-border bg-surface p-3 shadow-xl">
              <NavItems onNavigate={closeOnMobile} />
            </aside>
          </div>
        ) : null}

        <main id="main" className="min-w-0 flex-1 overflow-y-auto px-4 py-6 md:px-8 md:py-8">
          <div className="mx-auto w-full max-w-6xl">
            <Suspense fallback={<PageFallback />}>
              <div key={location.pathname} className="animate-fade-in-up">
                <Outlet />
              </div>
            </Suspense>
          </div>
        </main>
      </div>
    </div>
  );
}
