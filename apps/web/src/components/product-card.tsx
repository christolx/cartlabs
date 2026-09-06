"use client";

import Image from "next/image";
import Link from "next/link";
import { catalogNavigationKey } from "@/components/catalog-navigation";
import type { components } from "@/lib/api/schema";
import { formatMoney } from "@/lib/api/client";

type Product = components["schemas"]["ProductSummary"];

const editorialImages: Record<string, string> = {
  "arc-task-lamp": "/images/catalog-home/arc-task-lamp-editorial.webp",
  "canvas-tote": "/images/catalog-home/canvas-tote-editorial.webp",
  "compact-digital-camera":
    "/images/catalog-home/compact-digital-camera-editorial.webp",
  "field-bottle": "/images/catalog-home/field-bottle-editorial.webp",
  "handwoven-market-basket":
    "/images/catalog-home/handwoven-market-basket-editorial.webp",
  "portable-radio": "/images/catalog-home/portable-radio-editorial.webp",
  "stoneware-dinner-set":
    "/images/catalog-home/stoneware-dinner-set-editorial.webp",
  "waxed-utility-jacket":
    "/images/catalog-home/waxed-utility-jacket-editorial.webp",
};

export function ProductCard({
  product,
  position,
  editorial = false,
  returnTo,
  onProductNavigate,
}: {
  product: Product;
  position?: number;
  editorial?: boolean;
  returnTo?: string;
  onProductNavigate?: () => void;
}) {
  const imageUrl = editorialImages[product.slug] ?? product.imageUrl;
  const productPath = `/products/${product.slug}`;
  const productHref = returnTo
    ? `${productPath}?returnTo=${encodeURIComponent(returnTo)}`
    : productPath;

  function handleProductNavigate() {
    onProductNavigate?.();
    if (!returnTo) return;
    try {
      window.sessionStorage.setItem(
        catalogNavigationKey,
        JSON.stringify({ productPath, returnTo }),
      );
    } catch {
      // Product navigation still works when storage is unavailable or full.
    }
  }

  return (
    <article className={`product-card${editorial ? " editorial-card" : ""}`}>
      <Link
        href={productHref}
        className="product-image-link"
        aria-label={`View ${product.name}`}
        onNavigate={handleProductNavigate}
      >
        <div className="product-image">
          {position ? (
            <span className="product-number" aria-hidden="true">
              {String(position).padStart(2, "0")}
            </span>
          ) : null}
          {imageUrl ? (
            <Image
              src={imageUrl}
              alt=""
              fill
              sizes="(max-width: 767px) 100vw, (max-width: 1199px) 50vw, 25vw"
            />
          ) : (
            <div className="image-fallback">Image pending</div>
          )}
        </div>
      </Link>
      <div className="product-copy">
        <h3>
          <Link href={productHref} onNavigate={handleProductNavigate}>
            {product.name}
          </Link>
        </h3>
        <p className="product-store">
          <Link href={`/stores/${product.storeSlug}`}>{product.storeName}</Link>
        </p>
        <div className="product-meta">
          <span>{formatMoney(product.minPriceMinor, product.currency)}</span>
          {product.inStock ? (
            <Link
              className="stock-good"
              href={productHref}
              onNavigate={handleProductNavigate}
            >
              In stock&nbsp; →
            </Link>
          ) : (
            <span className="stock-empty">Sold out</span>
          )}
        </div>
      </div>
    </article>
  );
}
