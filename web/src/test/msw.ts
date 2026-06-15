import { graphql, HttpResponse } from "msw";
import { setupServer } from "msw/node";

import { GRAPHQL_URL } from "@/lib/env";

/**
 * Default handlers. Feature tests extend the server with `server.use(...)`
 * to override individual operations.
 */
export const handlers = [
  graphql.query("Health", () => {
    return HttpResponse.json({ data: { health: "ok" } });
  }),
];

export const server = setupServer(...handlers);

/** Re-exported so tests can build their own handlers against the same endpoint. */
export { graphql, HttpResponse, GRAPHQL_URL };
