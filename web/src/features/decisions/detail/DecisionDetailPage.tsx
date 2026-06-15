import { ArrowLeft, ChevronDown, Pencil } from "lucide-react";
import { useState } from "react";
import { Link, useParams } from "react-router-dom";

import { Badge } from "@/components/ui/badge";
import { Button, buttonVariants } from "@/components/ui/button";
import { Card, CardHeader, CardTitle } from "@/components/ui/card";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { EmptyState } from "@/components/ui/empty-state";
import { Skeleton } from "@/components/ui/skeleton";
import { useToast } from "@/components/ui/toaster";
import { STATUS_META } from "@/features/decisions/status";
import {
  type DecisionStatus,
  useDecisionQuery,
  useTransitionDecisionMutation,
} from "@/graphql/generated/graphql";
import { formatDate } from "@/lib/format";

import { AssumptionsSection } from "./AssumptionsSection";
import { ConfirmDialog } from "./ConfirmDialog";
import { EditDecisionDialog } from "./EditDecisionDialog";
import { OutcomeSection } from "./OutcomeSection";
import { PredictionsSection } from "./PredictionsSection";
import { allowedTransitions } from "./transitions";

function DetailSkeleton() {
  return (
    <div className="flex flex-col gap-6">
      <Skeleton className="h-5 w-32" />
      <Card>
        <CardHeader className="gap-3">
          <Skeleton className="h-7 w-2/3" />
          <Skeleton className="h-4 w-full" />
          <Skeleton className="h-4 w-1/2" />
        </CardHeader>
      </Card>
      <Skeleton className="h-48 w-full rounded-lg" />
      <Skeleton className="h-48 w-full rounded-lg" />
    </div>
  );
}

export function DecisionDetailPage() {
  const { id } = useParams<{ id: string }>();
  const toast = useToast();

  const [editOpen, setEditOpen] = useState(false);
  const [pendingStatus, setPendingStatus] = useState<DecisionStatus | null>(
    null,
  );

  const { data, loading, error, refetch } = useDecisionQuery({
    variables: { id: id ?? "" },
    skip: !id,
  });

  const [transitionDecision, { loading: transitioning }] =
    useTransitionDecisionMutation();

  const decision = data?.decision;

  async function runTransition(status: DecisionStatus) {
    if (!decision) {
      return;
    }
    try {
      const { data: result } = await transitionDecision({
        variables: { input: { id: decision.id, status } },
      });
      const payload = result?.transitionDecision;
      if (payload && payload.userErrors.length > 0) {
        toast.error(payload.userErrors[0].message);
        return;
      }
      toast.success(`Moved to ${STATUS_META[status].label}`);
      setPendingStatus(null);
      await refetch();
    } catch {
      toast.error("Something went wrong. Please try again.");
    }
  }

  function handleTransitionSelect(status: DecisionStatus) {
    if (status === "ARCHIVED") {
      setPendingStatus(status);
      return;
    }
    void runTransition(status);
  }

  if (loading) {
    return <DetailSkeleton />;
  }

  if (error || !decision) {
    return (
      <div className="flex flex-col gap-6">
        <BackLink />
        <EmptyState
          title="Decision not found"
          description="This decision may have been removed, or you may not have access to it."
          action={
            <Link to="/decisions" className={buttonVariants()}>
              Back to decisions
            </Link>
          }
        />
      </div>
    );
  }

  const status = STATUS_META[decision.status];
  const nextStatuses = allowedTransitions(decision.status);

  return (
    <div className="flex flex-col gap-6">
      <BackLink />

      <Card>
        <CardHeader className="gap-4">
          <div className="flex flex-wrap items-start justify-between gap-4">
            <div className="flex min-w-0 flex-col gap-2">
              <div className="flex flex-wrap items-center gap-3">
                <CardTitle className="text-2xl">{decision.title}</CardTitle>
                <Badge variant={status.variant}>{status.label}</Badge>
              </div>
              {decision.description ? (
                <p className="whitespace-pre-wrap break-words text-sm text-muted-foreground">
                  {decision.description}
                </p>
              ) : null}
            </div>

            <div className="flex shrink-0 items-center gap-2">
              <Button
                variant="outline"
                size="sm"
                onClick={() => setEditOpen(true)}
              >
                <Pencil className="size-4" />
                Edit
              </Button>

              {nextStatuses.length > 0 ? (
                <DropdownMenu>
                  <DropdownMenuTrigger asChild>
                    <Button size="sm" loading={transitioning}>
                      Move to
                      <ChevronDown className="size-4" />
                    </Button>
                  </DropdownMenuTrigger>
                  <DropdownMenuContent align="end">
                    {nextStatuses.map((next) => (
                      <DropdownMenuItem
                        key={next}
                        onSelect={() => handleTransitionSelect(next)}
                      >
                        {STATUS_META[next].label}
                      </DropdownMenuItem>
                    ))}
                  </DropdownMenuContent>
                </DropdownMenu>
              ) : null}
            </div>
          </div>

          <dl className="flex flex-wrap gap-x-8 gap-y-2 text-sm">
            <div className="flex gap-2">
              <dt className="text-muted-foreground">Owner</dt>
              <dd className="font-medium text-foreground">
                {decision.owner.name}
              </dd>
            </div>
            <div className="flex gap-2">
              <dt className="text-muted-foreground">Created</dt>
              <dd className="font-medium tabular-nums text-foreground">
                {formatDate(decision.createdAt)}
              </dd>
            </div>
            {decision.decidedAt ? (
              <div className="flex gap-2">
                <dt className="text-muted-foreground">Decided</dt>
                <dd className="font-medium tabular-nums text-foreground">
                  {formatDate(decision.decidedAt)}
                </dd>
              </div>
            ) : null}
          </dl>
        </CardHeader>
      </Card>

      <AssumptionsSection
        decisionId={decision.id}
        assumptions={decision.assumptions}
        onChanged={() => void refetch()}
      />

      <PredictionsSection
        decisionId={decision.id}
        predictions={decision.predictions}
        onChanged={() => void refetch()}
      />

      <OutcomeSection
        decisionId={decision.id}
        status={decision.status}
        outcome={decision.outcome}
        onChanged={() => void refetch()}
      />

      <EditDecisionDialog
        decisionId={decision.id}
        initialTitle={decision.title}
        initialDescription={decision.description}
        open={editOpen}
        onOpenChange={setEditOpen}
        onSaved={() => void refetch()}
      />

      <ConfirmDialog
        open={pendingStatus === "ARCHIVED"}
        onOpenChange={(open) => {
          if (!open) {
            setPendingStatus(null);
          }
        }}
        title="Archive this decision?"
        description="Archived decisions are read-only and drop out of the active list. You can still view their history."
        confirmLabel="Archive"
        destructive
        loading={transitioning}
        onConfirm={() => void runTransition("ARCHIVED")}
      />
    </div>
  );
}

function BackLink() {
  return (
    <Link
      to="/decisions"
      className="inline-flex w-fit items-center gap-1.5 text-sm text-muted-foreground transition-colors hover:text-foreground"
    >
      <ArrowLeft className="size-4" />
      Back to decisions
    </Link>
  );
}
