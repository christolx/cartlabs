import { expect, it, vi } from "vitest";
import { fireEvent, render, screen } from "@testing-library/react";
import { EmptyState, ErrorState, LoadingState } from "./async-state";

it("exposes accessible loading state", () => {
  render(<LoadingState label="Loading purchases" />);
  const label = screen.getByText("Loading purchases");
  expect(label.parentElement?.getAttribute("aria-busy")).toBe("true");
});

it("renders useful empty action", () => {
  render(
    <EmptyState
      title="No purchases"
      message="Browse first."
      href="/"
      action="Browse catalog"
    />,
  );
  expect(
    screen.getByRole("link", { name: "Browse catalog" }).getAttribute("href"),
  ).toBe("/");
});

it("retries contextual errors", () => {
  const retry = vi.fn();
  render(<ErrorState message="API unavailable" retry={retry} />);
  fireEvent.click(screen.getByRole("button", { name: "Try again" }));
  expect(retry).toHaveBeenCalledOnce();
});
