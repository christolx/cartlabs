import Image from "next/image";
import Link from "next/link";
import type { components } from "@/lib/api/schema";
import { formatMoney } from "@/lib/api/client";

type Product = components["schemas"]["ProductSummary"];

export function ProductCard({ product }: { product: Product }) {
  return (
    <article className="product-card">
      <Link href={`/products/${product.slug}`} className="product-image-link" aria-label={`View ${product.name}`}>
        <div className="product-image">
          {product.imageUrl ? (
            <Image
              src={product.imageUrl}
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
        <p className="product-store">{product.storeName}</p>
        <h3>
          <Link href={`/products/${product.slug}`}>{product.name}</Link>
        </h3>
        <div className="product-meta">
          <span>{formatMoney(product.minPriceMinor, product.currency)}</span>
          <span className={product.inStock ? "stock-good" : "stock-empty"}>
            {product.inStock ? "In stock" : "Sold out"}
          </span>
        </div>
      </div>
    </article>
  );
}
