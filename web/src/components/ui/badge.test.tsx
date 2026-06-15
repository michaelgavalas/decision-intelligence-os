import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";

import { Badge } from "@/components/ui/badge";

describe("Badge", () => {
  it("renders its label", () => {
    render(<Badge>Draft</Badge>);
    expect(screen.getByText("Draft")).toBeInTheDocument();
  });

  it("applies status variant styling", () => {
    render(<Badge variant="decided">Decided</Badge>);
    const badge = screen.getByText("Decided");
    expect(badge.className).toContain("text-success");
  });

  it("applies the archived variant with a strike-through", () => {
    render(<Badge variant="archived">Archived</Badge>);
    expect(screen.getByText("Archived").className).toContain("line-through");
  });
});
