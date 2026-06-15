import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { MemoryRouter, Route, Routes } from "react-router-dom";

import { NotFoundPage } from "@/pages/not-found-page";

describe("NotFoundPage", () => {
  it("renders for an unknown route", () => {
    render(
      <MemoryRouter initialEntries={["/this/does/not/exist"]}>
        <Routes>
          <Route path="/decisions" element={<div>Decisions screen</div>} />
          <Route path="*" element={<NotFoundPage />} />
        </Routes>
      </MemoryRouter>,
    );

    expect(screen.getByText("Page not found")).toBeInTheDocument();
    expect(
      screen.getByRole("link", { name: /back to decisions/i }),
    ).toHaveAttribute("href", "/decisions");
  });
});
