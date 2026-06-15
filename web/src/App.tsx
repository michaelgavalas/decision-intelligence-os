import { ApolloProvider } from "@apollo/client";
import { RouterProvider } from "react-router-dom";

import { ErrorBoundary } from "@/components/error-boundary";
import { Toaster } from "@/components/ui/toaster";
import { TooltipProvider } from "@/components/ui/tooltip";
import { apolloClient } from "@/lib/apollo";
import { router } from "@/router";

export function App() {
  return (
    <ErrorBoundary>
      <ApolloProvider client={apolloClient}>
        <TooltipProvider delayDuration={200}>
          <RouterProvider router={router} />
          <Toaster />
        </TooltipProvider>
      </ApolloProvider>
    </ErrorBoundary>
  );
}
