import { ListChecks, Plus } from "lucide-react";
import { useState } from "react";
import { Link } from "react-router-dom";

import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import { EmptyState } from "@/components/ui/empty-state";
import { Skeleton } from "@/components/ui/skeleton";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { CreateDecisionDialog } from "@/features/decisions/CreateDecisionDialog";
import { STATUS_META } from "@/features/decisions/status";
import { useCurrentTeam } from "@/features/teams/use-current-team";
import { useDecisionsQuery } from "@/graphql/generated/graphql";
import { formatDate } from "@/lib/format";

const PAGE_SIZE = 20;
const SKELETON_ROWS = 5;

function DecisionsTableSkeleton() {
  return (
    <Table>
      <TableHeader>
        <TableRow>
          <TableHead>Title</TableHead>
          <TableHead>Status</TableHead>
          <TableHead>Owner</TableHead>
          <TableHead>Created</TableHead>
          <TableHead>Decided</TableHead>
        </TableRow>
      </TableHeader>
      <TableBody>
        {Array.from({ length: SKELETON_ROWS }).map((_, index) => (
          <TableRow key={index}>
            {Array.from({ length: 5 }).map((__, cell) => (
              <TableCell key={cell}>
                <Skeleton className="h-4 w-full max-w-32" />
              </TableCell>
            ))}
          </TableRow>
        ))}
      </TableBody>
    </Table>
  );
}

export function DecisionsListPage() {
  const { currentTeamId, loading: teamLoading } = useCurrentTeam();
  const [dialogOpen, setDialogOpen] = useState(false);

  const { data, loading, error, fetchMore, refetch } = useDecisionsQuery({
    variables: { teamId: currentTeamId ?? "", first: PAGE_SIZE },
    skip: !currentTeamId,
    notifyOnNetworkStatusChange: true,
  });

  const [loadingMore, setLoadingMore] = useState(false);

  const connection = data?.decisions;
  const edges = connection?.edges ?? [];
  const pageInfo = connection?.pageInfo;
  const totalCount = connection?.totalCount ?? 0;

  async function handleLoadMore() {
    if (!pageInfo?.endCursor || !currentTeamId) {
      return;
    }
    setLoadingMore(true);
    try {
      await fetchMore({
        variables: {
          teamId: currentTeamId,
          first: PAGE_SIZE,
          after: pageInfo.endCursor,
        },
      });
    } finally {
      setLoadingMore(false);
    }
  }

  const showInitialLoading =
    (teamLoading && !currentTeamId) || (loading && edges.length === 0 && !error);

  return (
    <div className="flex flex-col gap-6">
      <header className="flex flex-wrap items-end justify-between gap-4">
        <div className="flex flex-col gap-1">
          <h1 className="text-2xl font-bold tracking-tight text-foreground">
            Decisions
          </h1>
          {totalCount > 0 ? (
            <p className="text-sm text-muted-foreground">
              {totalCount} {totalCount === 1 ? "decision" : "decisions"}
            </p>
          ) : null}
        </div>
        <Button onClick={() => setDialogOpen(true)} disabled={!currentTeamId}>
          <Plus className="size-4" />
          New decision
        </Button>
      </header>

      {showInitialLoading ? (
        <DecisionsTableSkeleton />
      ) : error ? (
        <Card className="flex flex-col items-center gap-3 p-8 text-center">
          <p className="text-sm font-medium text-foreground">
            We couldn&apos;t load decisions.
          </p>
          <p className="text-sm text-muted-foreground">
            Something went wrong while fetching this team&apos;s decisions.
          </p>
          <Button variant="outline" onClick={() => void refetch()}>
            Try again
          </Button>
        </Card>
      ) : edges.length === 0 ? (
        <EmptyState
          icon={ListChecks}
          title="No decisions yet"
          description="Start tracking your team's decisions from draft to recorded outcome."
          action={
            <Button onClick={() => setDialogOpen(true)} disabled={!currentTeamId}>
              <Plus className="size-4" />
              Create your first decision
            </Button>
          }
        />
      ) : (
        <div className="flex flex-col gap-4">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Title</TableHead>
                <TableHead>Status</TableHead>
                <TableHead>Owner</TableHead>
                <TableHead>Created</TableHead>
                <TableHead>Decided</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {edges.map(({ node }) => {
                const status = STATUS_META[node.status];
                return (
                  <TableRow key={node.id}>
                    <TableCell className="font-medium">
                      <Link
                        to={`/decisions/${node.id}`}
                        className="text-foreground hover:text-primary hover:underline"
                      >
                        {node.title}
                      </Link>
                    </TableCell>
                    <TableCell>
                      <Badge variant={status.variant}>{status.label}</Badge>
                    </TableCell>
                    <TableCell className="text-muted-foreground">
                      {node.owner.name}
                    </TableCell>
                    <TableCell className="font-mono text-muted-foreground tabular-nums">
                      {formatDate(node.createdAt)}
                    </TableCell>
                    <TableCell className="font-mono text-muted-foreground tabular-nums">
                      {formatDate(node.decidedAt)}
                    </TableCell>
                  </TableRow>
                );
              })}
            </TableBody>
          </Table>

          {pageInfo?.hasNextPage ? (
            <div className="flex justify-center">
              <Button
                variant="outline"
                onClick={() => void handleLoadMore()}
                loading={loadingMore}
                disabled={loadingMore}
              >
                Load more
              </Button>
            </div>
          ) : null}
        </div>
      )}

      {currentTeamId ? (
        <CreateDecisionDialog
          teamId={currentTeamId}
          open={dialogOpen}
          onOpenChange={setDialogOpen}
        />
      ) : null}
    </div>
  );
}
