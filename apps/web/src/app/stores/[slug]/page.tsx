import type { Metadata } from "next";
import Link from "next/link";
import { notFound } from "next/navigation";
import { ProductCard } from "@/components/product-card";
import {
  APIError,
  apiGet,
  type Category,
  type ProductPage,
} from "@/lib/api/client";
import type { components } from "@/lib/api/schema";

type Store = components["schemas"]["StoreProfile"];
type Search = {
  q?: string;
  category?: string;
  minPrice?: string;
  maxPrice?: string;
  inStock?: string;
  page?: string;
};

export async function generateMetadata({
  params,
}: PageProps<"/stores/[slug]">): Promise<Metadata> {
  const { slug } = await params;
  try {
    const store = await apiGet<Store>(
      `/catalog/stores/${encodeURIComponent(slug)}`,
    );
    return { title: store.name, description: store.description.slice(0, 155) };
  } catch {
    return { title: "Store | Cartlabs" };
  }
}

function parsePositive(value: string | undefined, fallback: number) {
  if (!value) return fallback;
  const parsed = Number(value);
  return Number.isSafeInteger(parsed) && parsed >= 0 ? parsed : Number.NaN;
}

function storePageHref(slug: string, filters: Search, page: number) {
  const query = new URLSearchParams();
  for (const [key, value] of Object.entries(filters))
    if (value && key !== "page") query.set(key, value);
  query.set("page", String(page));
  return `/stores/${slug}?${query.toString()}#store-catalog`;
}

export default async function StorePage({
  params,
  searchParams,
}: PageProps<"/stores/[slug]"> & { searchParams: Promise<Search> }) {
  const [{ slug }, filters] = await Promise.all([params, searchParams]);
  let store: Store;
  try {
    store = await apiGet<Store>(`/catalog/stores/${encodeURIComponent(slug)}`);
  } catch (error) {
    if (error instanceof APIError && error.status === 404) notFound();
    throw error;
  }
  const minPrice = parsePositive(filters.minPrice, 0);
  const maxPrice = filters.maxPrice
    ? parsePositive(filters.maxPrice, 0)
    : undefined;
  const page = parsePositive(filters.page, 1);
  const invalid =
    Number.isNaN(minPrice) ||
    Number.isNaN(page) ||
    page < 1 ||
    (maxPrice !== undefined && (Number.isNaN(maxPrice) || maxPrice < minPrice));
  const query = new URLSearchParams({ store: slug, pageSize: "20" });
  if (filters.q?.trim()) query.set("q", filters.q.trim());
  if (filters.category) query.set("category", filters.category);
  if (minPrice > 0) query.set("minPrice", String(minPrice));
  if (maxPrice !== undefined && !Number.isNaN(maxPrice))
    query.set("maxPrice", String(maxPrice));
  if (filters.inStock === "true") query.set("inStock", "true");
  if (page > 1) query.set("page", String(page));
  let catalog: ProductPage | null = null;
  let categories: Category[] = [];
  let failed = false;
  if (!invalid) {
    try {
      [catalog, { items: categories }] = await Promise.all([
        apiGet<ProductPage>(`/catalog/products?${query}`),
        apiGet<{ items: Category[] }>("/categories"),
      ]);
    } catch {
      failed = true;
    }
  }
  const pages = catalog
    ? Math.max(1, Math.ceil(catalog.total / catalog.pageSize))
    : 1;
  return (
    <main className="store-page shell">
      <section className="store-profile">
        <p className="eyebrow">Verified store</p>
        <h1>{store.name}</h1>
        <p>{store.description}</p>
        <a className="store-browse-action" href="#store-catalog">
          Browse store ↓
        </a>
        <dl>
          <div>
            <dt>Seller</dt>
            <dd>{store.sellerDisplayName}</dd>
          </div>
          <div>
            <dt>On Cartlabs since</dt>
            <dd>{new Date(store.createdAt).toLocaleDateString("id-ID")}</dd>
          </div>
          <div>
            <dt>Published products</dt>
            <dd>{catalog?.total ?? "—"}</dd>
          </div>
        </dl>
      </section>
      <section id="store-catalog" className="catalog">
        <div className="section-heading">
          <h2>Products from {store.name}</h2>
          <p>
            {catalog ? `${catalog.total} published products` : "Store catalog"}
          </p>
        </div>
        <form className="filters" method="get">
          <div className="field search-field">
            <label htmlFor="q">Search this store</label>
            <input id="q" name="q" defaultValue={filters.q} maxLength={100} />
          </div>
          <div className="field">
            <label htmlFor="category">Category</label>
            <select
              id="category"
              name="category"
              defaultValue={filters.category ?? ""}
            >
              <option value="">All categories</option>
              {categories.map((category) => (
                <option key={category.id} value={category.slug}>
                  {category.name}
                </option>
              ))}
            </select>
          </div>
          <div className="field">
            <label htmlFor="minPrice">Minimum price</label>
            <input
              id="minPrice"
              name="minPrice"
              type="number"
              min="0"
              defaultValue={filters.minPrice}
            />
          </div>
          <div className="field">
            <label htmlFor="maxPrice">Maximum price</label>
            <input
              id="maxPrice"
              name="maxPrice"
              type="number"
              min="0"
              defaultValue={filters.maxPrice}
            />
          </div>
          <label className="check-field">
            <input
              type="checkbox"
              name="inStock"
              value="true"
              defaultChecked={filters.inStock === "true"}
            />
            In stock only
          </label>
          <div className="filter-actions">
            <button className="button button-primary">Apply filters</button>
            <Link
              className="button button-secondary"
              href={`/stores/${slug}#store-catalog`}
            >
              Clear
            </Link>
          </div>
        </form>
        {invalid ? (
          <div className="state-panel" role="alert">
            <h2>Invalid filters</h2>
            <p>Check price range and page.</p>
          </div>
        ) : failed ? (
          <div className="state-panel" role="alert">
            <h2>Store catalog unavailable</h2>
            <p>Try again after service connection returns.</p>
          </div>
        ) : catalog?.items.length ? (
          <>
            <div className="product-grid">
              {catalog.items.map((product) => (
                <ProductCard key={product.id} product={product} editorial />
              ))}
            </div>
            <nav className="pagination" aria-label="Store catalog pages">
              <Link
                className="button button-secondary"
                aria-disabled={catalog.page <= 1}
                href={storePageHref(
                  slug,
                  filters,
                  Math.max(1, catalog.page - 1),
                )}
              >
                Previous
              </Link>
              <span>
                Page {catalog.page} of {pages}
              </span>
              <Link
                className="button button-secondary"
                aria-disabled={catalog.page >= pages}
                href={storePageHref(
                  slug,
                  filters,
                  Math.min(pages, catalog.page + 1),
                )}
              >
                Next
              </Link>
            </nav>
          </>
        ) : (
          <div className="state-panel">
            <h2>
              {Object.values(filters).some(Boolean)
                ? "No products match"
                : "No published products"}
            </h2>
            <p>
              {Object.values(filters).some(Boolean)
                ? "Clear filters to see more from this store."
                : "Store has no published products yet."}
            </p>
          </div>
        )}
      </section>
    </main>
  );
}
