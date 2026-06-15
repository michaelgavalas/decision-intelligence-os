import { Link } from "react-router-dom";
import { Compass } from "lucide-react";

import { buttonVariants } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";

export function NotFoundPage() {
  return (
    <div className="flex min-h-dvh items-center justify-center bg-background px-4">
      <Card className="w-full max-w-md text-center">
        <CardHeader className="items-center gap-3">
          <div className="flex size-12 items-center justify-center rounded-full bg-muted text-muted-foreground">
            <Compass className="size-6" />
          </div>
          <CardTitle>Page not found</CardTitle>
          <CardDescription>
            The page you are looking for doesn&apos;t exist or has moved.
          </CardDescription>
        </CardHeader>
        <CardContent />
        <CardFooter className="justify-center">
          <Link to="/decisions" className={buttonVariants()}>
            Back to decisions
          </Link>
        </CardFooter>
      </Card>
    </div>
  );
}
