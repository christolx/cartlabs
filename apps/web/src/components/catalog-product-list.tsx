"use client";

import { useEffect, useEffectEvent, useRef, useState } from "react";
import { ProductCard } from "@/components/product-card";
import {
  catalogDepthKey,
  catalogDepthResetEvent,
} from "@/components/catalog-depth";
import type { ProductPage } from "@/lib/api/client";

type Product = ProductPage["items"][number];
const catalogDocumentKey = "cartlabs:catalog-document";
const catalogStatePrefix = "cartlabs:catalog-state:";

type CatalogState = {
  products: Product[];
  hasMore: boolean;
  scrollY: number;
};

function catalogStateKey(catalogKey: string) {
  return `${catalogStatePrefix}${catalogKey}`;
}

function clearCatalogStates() {
  for (let index = window.sessionStorage.length - 1; index >= 0; index -= 1) {
    const key = window.sessionStorage.key(index);
    if (key?.startsWith(catalogStatePrefix))
      window.sessionStorage.removeItem(key);
  }
}

function readCatalogState(catalogKey: string): CatalogState | null {
  try {
    const value = window.sessionStorage.getItem(catalogStateKey(catalogKey));
    if (!value) return null;
    const state = JSON.parse(value) as Partial<CatalogState>;
    if (
      !Array.isArray(state.products) ||
      state.products.length === 0 ||
      state.products.length > 1_000 ||
      typeof state.hasMore !== "boolean" ||
      typeof state.scrollY !== "number" ||
      !Number.isFinite(state.scrollY)
    )
      return null;
    return state as CatalogState;
  } catch {
    return null;
  }
}

function writeCatalogState(catalogKey: string, state: CatalogState) {
  try {
    window.sessionStorage.setItem(
      catalogStateKey(catalogKey),
      JSON.stringify(state),
    );
  } catch {
    // Catalog still works when storage is unavailable or full.
  }
}

function restoreScrollPosition(scrollY: number, attempt = 0) {
  window.requestAnimationFrame(() => {
    window.scrollTo(0, scrollY);
    if (Math.abs(window.scrollY - scrollY) > 1 && attempt < 20) {
      window.setTimeout(() => restoreScrollPosition(scrollY, attempt + 1), 50);
    }
  });
}

function initializeCatalogDepth() {
  const documentId = String(window.performance.timeOrigin);
  if (window.sessionStorage.getItem(catalogDocumentKey) === documentId) return;

  const navigation = window.performance.getEntriesByType("navigation")[0] as
    PerformanceNavigationTiming | undefined;
  if (navigation?.type === "reload") {
    window.sessionStorage.removeItem(catalogDepthKey);
    clearCatalogStates();
  }
  window.sessionStorage.setItem(catalogDocumentKey, documentId);
}

export function CatalogProductList({
  initialProducts,
  initialPosition,
  initialHasMore,
  filterQuery,
  catalogKey,
  returnTo,
  loadPage,
}: {
  initialProducts: Product[];
  initialPosition: number;
  initialHasMore: boolean;
  filterQuery: string;
  catalogKey: string;
  returnTo: string;
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
  const activeCatalogKey = useRef(catalogKey);
  const resetGeneration = useRef(0);

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

  const resetCatalog = useEffectEvent(() => {
    resetGeneration.current += 1;
    window.sessionStorage.removeItem(catalogDepthKey);
    clearCatalogStates();
    visibleCount.current = initialProducts.length;
    setProducts(initialProducts);
    setHasMore(initialHasMore);
    setLoading(false);
    setError("");
  });

  const startRebuild = useEffectEvent((isCancelled: () => boolean) => {
    initializeCatalogDepth();
    if (activeCatalogKey.current !== catalogKey) {
      activeCatalogKey.current = catalogKey;
      visibleCount.current = initialProducts.length;
      setProducts(initialProducts);
      setHasMore(initialHasMore);
      setError("");
    }
    const savedState = readCatalogState(catalogKey);
    if (savedState && savedState.products.length >= initialProducts.length) {
      visibleCount.current = savedState.products.length;
      setProducts(savedState.products);
      setHasMore(savedState.hasMore);
      setError("");
      setLoading(false);
      window.sessionStorage.setItem(
        catalogDepthKey,
        String(savedState.products.length),
      );
      restoreScrollPosition(savedState.scrollY);
      return;
    }
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
    if (visibleCount.current >= targetCount) return;
    void rebuildCatalog(targetCount, isCancelled);
  });

  useEffect(() => {
    const handleReset = () => resetCatalog();
    window.addEventListener(catalogDepthResetEvent, handleReset);
    return () =>
      window.removeEventListener(catalogDepthResetEvent, handleReset);
  }, []);

  useEffect(() => {
    let cancelled = false;
    const generation = resetGeneration.current;
    startRebuild(() => cancelled || resetGeneration.current !== generation);
    return () => {
      cancelled = true;
    };
  }, [catalogKey, filterQuery]);

  async function loadMore() {
    if (loading || !hasMore) return;
    const generation = resetGeneration.current;
    setLoading(true);
    setError("");
    try {
      const nextPage = await loadPage(
        filterQuery,
        products.map((product) => product.id),
      );
      if (resetGeneration.current !== generation) return;
      const updated = [...products, ...nextPage.items];
      visibleCount.current = updated.length;
      window.sessionStorage.setItem(catalogDepthKey, String(updated.length));
      writeCatalogState(catalogKey, {
        products: updated,
        hasMore: nextPage.hasMore,
        scrollY: window.scrollY,
      });
      setProducts(updated);
      setHasMore(nextPage.hasMore);
    } catch {
      if (resetGeneration.current === generation)
        setError("More products could not be loaded. Try again.");
    } finally {
      if (resetGeneration.current === generation) setLoading(false);
    }
  }

  function rememberCatalogPosition() {
    writeCatalogState(catalogKey, {
      products,
      hasMore,
      scrollY: window.scrollY,
    });
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
            returnTo={returnTo}
            onProductNavigate={rememberCatalogPosition}
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
