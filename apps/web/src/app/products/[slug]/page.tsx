import type { Metadata } from "next";
import Image from "next/image";
import Link from "next/link";
import { notFound } from "next/navigation";
import { AddToCart } from "@/components/add-to-cart";
import {
  APIError,
  apiGet,
  formatMoney,
  type ProductDetail,
} from "@/lib/api/client";
import type { components } from "@/lib/api/schema";

type ReviewSummary = components["schemas"]["ReviewSummary"];

export async function generateMetadata({
  params,
}: PageProps<"/products/[slug]">): Promise<Metadata> {
  const { slug } = await params;
  try {
    const product = await apiGet<ProductDetail>(
      `/catalog/products/${encodeURIComponent(slug)}`,
    );
    return {
      title: product.name,
      description: product.description.slice(0, 155),
    };
  } catch {
    return { title: "Product | Cartlabs" };
  }
}

export default async function ProductPage({
  params,
}: PageProps<"/products/[slug]">) {
  const { slug } = await params;
  let product: ProductDetail;
  let reviews: ReviewSummary;
  try {
    [product, reviews] = await Promise.all([
      apiGet<ProductDetail>(`/catalog/products/${encodeURIComponent(slug)}`),
      apiGet<ReviewSummary>(
        `/catalog/products/${encodeURIComponent(slug)}/reviews`,
      ),
    ]);
  } catch (error) {
    if (error instanceof APIError && error.status === 404) notFound();
    throw error;
  }
  const firstVariant = product.variants[0];
  const firstImage = product.images[0];
  const totalStock = product.variants.reduce(
    (total, variant) => total + variant.stock,
    0,
  );

  return (
    <main className="product-page shell">
      <Link className="back-link" href="/#catalog">
        Back to catalog
      </Link>
      <article className="product-detail">
        <div className="detail-gallery">
          <div className="detail-media">
            {firstImage ? (
              <Image
                src={firstImage.url}
                alt={firstImage.altText}
                fill
                priority
                sizes="(max-width: 767px) 100vw, 58vw"
              />
            ) : (
              <div className="image-fallback">Image pending</div>
            )}
            <span className="media-count">
              {product.images.length || 0} image
              {product.images.length === 1 ? "" : "s"}
            </span>
          </div>
          {product.images.length > 1 ? (
            <div className="detail-thumbnail-rail" aria-label="Product images">
              {product.images.map((image, index) => (
                <div className="detail-thumbnail" key={image.id}>
                  <Image src={image.url} alt="" width={96} height={72} />
                  <span>{String(index + 1).padStart(2, "0")}</span>
                </div>
              ))}
            </div>
          ) : null}
        </div>
        <div className="detail-copy">
          <p className="product-store">
            <Link className="text-link" href={`/stores/${product.storeSlug}`}>
              {product.storeName}
            </Link>{" "}
            / {product.category.name}
          </p>
          <h1>{product.name}</h1>
          <dl className="detail-facts">
            <div>
              <dt>Store</dt>
              <dd>{product.storeName}</dd>
            </div>
            <div>
              <dt>Category</dt>
              <dd>{product.category.name}</dd>
            </div>
            <div>
              <dt>Reference</dt>
              <dd>{product.slug}</dd>
            </div>
            <div>
              <dt>Stock</dt>
              <dd>{totalStock} units</dd>
            </div>
          </dl>
          <p className="detail-description">{product.description}</p>
          {firstVariant && (
            <p className="detail-price">
              From {formatMoney(firstVariant.priceMinor, firstVariant.currency)}
            </p>
          )}
          <div
            className="variant-list"
            role="list"
            aria-label="Available variants"
          >
            {product.variants.map((variant) => (
              <div className="variant-row" role="listitem" key={variant.id}>
                <div>
                  <strong>{variant.name}</strong>
                  <span>
                    {variant.sku}
                    {Object.keys(variant.attributes).length
                      ? ` / ${Object.entries(variant.attributes)
                          .map(([key, value]) => `${key}: ${value}`)
                          .join(", ")}`
                      : ""}
                  </span>
                </div>
                <div>
                  <strong>
                    {formatMoney(variant.priceMinor, variant.currency)}
                  </strong>
                  <span>
                    {variant.stock > 0
                      ? `${variant.stock} available`
                      : "Sold out"}
                  </span>
                </div>
              </div>
            ))}
          </div>
          <AddToCart product={product} />
          <div className="seller-trust-strip">
            <span className="seller-trust-label">Verified seller</span>
            <div className="seller-store-row">
              <Link
                className="seller-store-name"
                href={`/stores/${product.storeSlug}`}
              >
                {product.storeName}
              </Link>
              <Link
                className="seller-store-link"
                href={`/stores/${product.storeSlug}`}
              >
                View store →
              </Link>
            </div>
          </div>
        </div>
      </article>
      <section className="review-section" aria-labelledby="reviews-heading">
        <div className="review-heading">
          <div>
            <p className="eyebrow">Verified purchases</p>
            <h2 id="reviews-heading">Buyer reviews</h2>
          </div>
          <div
            className="review-score"
            role="group"
            aria-label={`${reviews.average.toFixed(1)} out of 5 from ${reviews.count} reviews`}
          >
            <strong>
              {reviews.count ? reviews.average.toFixed(1) : "No score"}
            </strong>
            <span>
              {reviews.count} {reviews.count === 1 ? "review" : "reviews"}
            </span>
          </div>
        </div>
        {reviews.items.length ? (
          <div className="public-review-list">
            {reviews.items.map((review) => (
              <article key={review.id}>
                <div className="public-review-meta">
                  <strong>{review.title}</strong>
                  <span>{review.rating} / 5</span>
                </div>
                <p>{review.body}</p>
                <footer>
                  <span>{review.buyerName}</span>
                  <span>Verified purchase</span>
                  <time dateTime={review.createdAt}>
                    {new Date(review.createdAt).toLocaleDateString("id-ID")}
                  </time>
                </footer>
              </article>
            ))}
          </div>
        ) : (
          <p className="empty-copy">No verified reviews yet.</p>
        )}
      </section>
    </main>
  );
}
