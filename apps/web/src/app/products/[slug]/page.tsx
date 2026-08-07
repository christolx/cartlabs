import type { Metadata } from "next";
import Image from "next/image";
import Link from "next/link";
import { notFound } from "next/navigation";
import { SiteHeader } from "@/components/site-header";
import { APIError, apiGet, formatMoney, type ProductDetail } from "@/lib/api/client";
import type { components } from "@/lib/api/schema";

type ReviewSummary = components["schemas"]["ReviewSummary"];

export async function generateMetadata({ params }: PageProps<"/products/[slug]">): Promise<Metadata> {
  const { slug } = await params;
  try {
    const product = await apiGet<ProductDetail>(`/catalog/products/${encodeURIComponent(slug)}`);
    return { title: product.name, description: product.description.slice(0, 155) };
  } catch { return { title: "Product | Cartlabs" }; }
}

export default async function ProductPage({ params }: PageProps<"/products/[slug]">) {
  const { slug } = await params;
  let product: ProductDetail;
  let reviews: ReviewSummary;
  try {
    [product, reviews] = await Promise.all([
      apiGet<ProductDetail>(`/catalog/products/${encodeURIComponent(slug)}`),
      apiGet<ReviewSummary>(`/catalog/products/${encodeURIComponent(slug)}/reviews`),
    ]);
  } catch (error) {
    if (error instanceof APIError && error.status === 404) notFound();
    throw error;
  }
  const firstVariant = product.variants[0];
  const firstImage = product.images[0];

  return (
    <>
      <SiteHeader />
      <main className="product-page shell">
        <Link className="back-link" href="/#catalog">Back to catalog</Link>
        <article className="product-detail">
          <div className="detail-media">
            {firstImage ? (
              <Image src={firstImage.url} alt={firstImage.altText} fill priority sizes="(max-width: 767px) 100vw, 58vw" />
            ) : <div className="image-fallback">Image pending</div>}
          </div>
          <div className="detail-copy">
            <p className="product-store">{product.storeName}</p>
            <h1>{product.name}</h1>
            <p className="detail-description">{product.description}</p>
            {firstVariant && <p className="detail-price">From {formatMoney(firstVariant.priceMinor, firstVariant.currency)}</p>}
            <div className="variant-list" role="list" aria-label="Available variants">
              {product.variants.map((variant) => (
                <div className="variant-row" role="listitem" key={variant.id}>
                  <div><strong>{variant.name}</strong><span>{variant.sku}</span></div>
                  <div><strong>{formatMoney(variant.priceMinor, variant.currency)}</strong><span>{variant.stock > 0 ? `${variant.stock} available` : "Sold out"}</span></div>
                </div>
              ))}
            </div>
            <p className="purchase-note">Catalog stock shown live. Checkout available in role demo.</p>
          </div>
        </article>
        <section className="review-section" aria-labelledby="reviews-heading">
          <div className="review-heading">
            <div>
              <p className="eyebrow">Verified purchases</p>
              <h2 id="reviews-heading">Buyer reviews</h2>
            </div>
            <div className="review-score" role="group" aria-label={`${reviews.average.toFixed(1)} out of 5 from ${reviews.count} reviews`}>
              <strong>{reviews.count ? reviews.average.toFixed(1) : "—"}</strong>
              <span>{reviews.count} {reviews.count === 1 ? "review" : "reviews"}</span>
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
                  <footer><span>{review.buyerName}</span><span>Verified purchase</span><time dateTime={review.createdAt}>{new Date(review.createdAt).toLocaleDateString("id-ID")}</time></footer>
                </article>
              ))}
            </div>
          ) : <p className="empty-copy">No verified reviews yet.</p>}
        </section>
      </main>
    </>
  );
}
