import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { beforeEach, expect, it, vi } from "vitest";
import {
  catalogDepthKey,
  catalogDepthResetEvent,
} from "@/components/catalog-depth";
import type { ProductPage } from "@/lib/api/client";
import { CatalogProductList } from "./catalog-product-list";

type Product = ProductPage["items"][number];

beforeEach(() => window.sessionStorage.clear());

function products(start: number, count: number): Product[] {
  return Array.from({ length: count }, (_, offset) => {
    const number = start + offset;
    return {
      id: String(number),
      name: `Product ${number}`,
      slug: `product-${number}`,
      storeName: "Test Store",
      storeSlug: "test-store",
      category: { id: "category", name: "Category", slug: "category" },
      minPriceMinor: 10_000,
      currency: "IDR",
      inStock: true,
      imageUrl: "",
    };
  });
}

it("appends eight products per click below current rows", async () => {
  const loadPage = vi
    .fn()
    .mockResolvedValueOnce({ items: products(9, 8), hasMore: true })
    .mockResolvedValueOnce({ items: products(17, 2), hasMore: false });

  const { container } = render(
    <CatalogProductList
      initialProducts={products(1, 8)}
      initialPosition={0}
      initialHasMore
      filterQuery="category=category"
      catalogKey="category=category"
      loadPage={loadPage}
    />,
  );

  fireEvent.click(screen.getByRole("button", { name: /load more/i }));
  await waitFor(() =>
    expect(container.querySelectorAll(".product-card")).toHaveLength(16),
  );
  expect(loadPage).toHaveBeenLastCalledWith(
    "category=category",
    products(1, 8).map((product) => product.id),
  );

  fireEvent.click(screen.getByRole("button", { name: /load more/i }));
  await waitFor(() =>
    expect(container.querySelectorAll(".product-card")).toHaveLength(18),
  );
  expect(loadPage).toHaveBeenLastCalledWith(
    "category=category",
    products(1, 16).map((product) => product.id),
  );
  expect(screen.queryByRole("button", { name: /load more/i })).toBeNull();
});

it("rebuilds changed catalogs in new order while preserving loaded depth", async () => {
  const loadPage = vi
    .fn()
    .mockResolvedValueOnce({ items: products(9, 8), hasMore: true })
    .mockResolvedValueOnce({ items: products(109, 8), hasMore: true });

  const { container, rerender } = render(
    <CatalogProductList
      initialProducts={products(1, 8)}
      initialPosition={0}
      initialHasMore
      filterQuery="category=category"
      catalogKey="category=category"
      loadPage={loadPage}
    />,
  );
  fireEvent.click(screen.getByRole("button", { name: /load more/i }));
  await waitFor(() =>
    expect(container.querySelectorAll(".product-card")).toHaveLength(16),
  );

  rerender(
    <CatalogProductList
      initialProducts={products(101, 8)}
      initialPosition={0}
      initialHasMore
      filterQuery=""
      catalogKey="default"
      loadPage={loadPage}
    />,
  );

  await waitFor(() =>
    expect(screen.getByRole("heading", { name: "Product 116" })).toBeTruthy(),
  );
  expect(container.querySelectorAll(".product-card")).toHaveLength(16);
  expect(screen.queryByRole("heading", { name: "Product 1" })).toBeNull();
  expect(loadPage).toHaveBeenLastCalledWith(
    "",
    products(101, 8).map((product) => product.id),
  );
});

it("resets loaded depth when clearing filters", async () => {
  const loadPage = vi
    .fn()
    .mockResolvedValueOnce({ items: products(9, 8), hasMore: true });

  const { container, rerender } = render(
    <CatalogProductList
      initialProducts={products(1, 8)}
      initialPosition={0}
      initialHasMore
      filterQuery="category=category"
      catalogKey="category=category"
      loadPage={loadPage}
    />,
  );
  fireEvent.click(screen.getByRole("button", { name: /load more/i }));
  await waitFor(() =>
    expect(container.querySelectorAll(".product-card")).toHaveLength(16),
  );
  expect(window.sessionStorage.getItem(catalogDepthKey)).toBe("16");

  window.dispatchEvent(new Event(catalogDepthResetEvent));

  rerender(
    <CatalogProductList
      initialProducts={products(101, 8)}
      initialPosition={0}
      initialHasMore
      filterQuery=""
      catalogKey="default"
      loadPage={loadPage}
    />,
  );

  await waitFor(() =>
    expect(container.querySelectorAll(".product-card")).toHaveLength(8),
  );
  expect(screen.getByRole("heading", { name: "Product 101" })).toBeTruthy();
  expect(screen.queryByRole("heading", { name: "Product 116" })).toBeNull();
  expect(window.sessionStorage.getItem(catalogDepthKey)).toBe("8");
  expect(loadPage).toHaveBeenCalledTimes(1);
});
