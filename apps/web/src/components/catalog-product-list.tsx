"use client";

import { useEffect, useEffectEvent, useRef, useState } from "react";
import { ProductCard } from "@/components/product-card";
import type { ProductPage } from "@/lib/api/client";

type Product = ProductPage["items"][number];
const catalogDepthKey = "cartlabs:catalog-depth";
const catalogDocumentKey = "cartlabs:catalog-document";

function initializeCatalogDepth() {
  const documentId = String(window.performance.timeOrigin);
  if (window.sessionStorage.getItem(catalogDocumentKey) === documentId) return;

  const navigation = window.performance.getEntriesByType("navigation")[0] as
    PerformanceNavigationTiming | undefined;
  if (navigation?.type === "reload")
    window.sessionStorage.removeItem(catalogDepthKey);
  window.sessionStorage.setItem(catalogDocumentKey, documentId);
}

export function CatalogProductList({
  initialProducts,
  initialPosition,
  initialHasMore,
  filterQuery,
  catalogKey,
  loadPage,
}: {
  initialProducts: Product[];
  initialPosition: number;
  initialHasMore: boolean;
  filterQuery: string;
  catalogKey: string;
  loadPage: (
    filterQuery: string,
    excludeIds: string[],
  ) => Promise<{ items: Product[]; hasMore: boolean }>;
}) {
  const [products, setProducts] = useState(initialProducts);
  const [hasMore, setHasMore] = useState(initialHasMore);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");
  const visibleCount = useRef(initialProducts.length);

  const rebuildCatalog = useEffectEvent(
    async (targetCount: number, isCancelled: () => boolean) => {
      let nextProducts = initialProducts;
      let nextHasMore = initialHasMore;

      setProducts(nextProducts);
      setHasMore(nextHasMore);
      setError("");
      setLoading(nextProducts.length < targetCount && nextHasMore);

      try {
        while (nextProducts.length < targetCount && nextHasMore) {
          const nextPage = await loadPage(
            filterQuery,
            nextProducts.map((product) => product.id),
          );
          if (isCancelled()) return;
          if (nextPage.items.length === 0) {
            nextHasMore = false;
            break;
          }
          nextProducts = [...nextProducts, ...nextPage.items].slice(
            0,
            targetCount,
          );
          nextHasMore = nextPage.hasMore;
        }
        if (!isCancelled()) {
          setProducts(nextProducts);
          visibleCount.current = nextProducts.length;
          setHasMore(nextHasMore);
        }
      } catch {
        if (!isCancelled())
          setError("Catalog could not be refreshed. Try again.");
      } finally {
        if (!isCancelled()) setLoading(false);
      }
    },
  );

  const startRebuild = useEffectEvent((isCancelled: () => boolean) => {
    initializeCatalogDepth();
    const storedCount = Number(window.sessionStorage.getItem(catalogDepthKey));
    const targetCount = Math.max(
      initialProducts.length,
      visibleCount.current,
      Number.isSafeInteger(storedCount) &&
        storedCount > 0 &&
        storedCount <= 1_000
        ? storedCount
        : 0,
    );
    window.sessionStorage.setItem(catalogDepthKey, String(targetCount));
    void rebuildCatalog(targetCount, isCancelled);
  });

  useEffect(() => {
    let cancelled = false;
    startRebuild(() => cancelled);
    return () => {
      cancelled = true;
    };
  }, [catalogKey]);

  async function loadMore() {
    if (loading || !hasMore) return;
    setLoading(true);
    setError("");
    try {
      const nextPage = await loadPage(
        filterQuery,
        products.map((product) => product.id),
      );
      setProducts((current) => {
        const updated = [...current, ...nextPage.items];
        visibleCount.current = updated.length;
        window.sessionStorage.setItem(catalogDepthKey, String(updated.length));
        return updated;
      });
      setHasMore(nextPage.hasMore);
    } catch {
      setError("More products could not be loaded. Try again.");
    } finally {
      setLoading(false);
    }
  }

  return (
    <>
      <div className="product-grid home-product-grid">
        {products.map((product, index) => (
          <ProductCard
            key={product.id}
            product={product}
            position={initialPosition + index + 1}
            editorial
          />
        ))}
      </div>
      {error ? (
        <p className="load-more-error" role="alert">
          {error}
        </p>
      ) : null}
      {hasMore ? (
        <nav className="pagination" aria-label="Catalog pages">
          <button
            className="button button-secondary load-more"
            type="button"
            disabled={loading}
            aria-busy={loading}
            onClick={() => void loadMore()}
          >
            {loading ? "Loading…" : "Load more"}{" "}
            <span aria-hidden="true">↓</span>
          </button>
        </nav>
      ) : null}
    </>
  );
}
