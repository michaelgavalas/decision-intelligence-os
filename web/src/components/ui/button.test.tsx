import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";

import { Button } from "@/components/ui/button";

describe("Button", () => {
  it("renders its children", () => {
    render(<Button>Save decision</Button>);
    expect(
      screen.getByRole("button", { name: "Save decision" }),
    ).toBeInTheDocument();
  });

  it("calls onClick when activated", async () => {
    const onClick = vi.fn();
    const user = userEvent.setup();
    render(<Button onClick={onClick}>Click me</Button>);

    await user.click(screen.getByRole("button", { name: "Click me" }));

    expect(onClick).toHaveBeenCalledOnce();
  });

  it("is disabled and shows a spinner while loading", async () => {
    const onClick = vi.fn();
    const user = userEvent.setup();
    render(
      <Button loading onClick={onClick}>
        Submitting
      </Button>,
    );

    const button = screen.getByRole("button", { name: /Submitting/ });
    expect(button).toBeDisabled();
    expect(button).toHaveAttribute("aria-busy", "true");
    expect(screen.getByRole("status")).toBeInTheDocument();

    await user.click(button);
    expect(onClick).not.toHaveBeenCalled();
  });
});
