import Image from "next/image";
import { loadCatalogPage } from "@/app/catalog-actions";
import { CatalogClearLink } from "@/components/catalog-clear-link";
import { CatalogProductList } from "@/components/catalog-product-list";
import {
  APIError,
  apiGet,
  type Category,
  type ProductPage,
} from "@/lib/api/client";

type Search = {
  q?: string;
  category?: string;
  minPrice?: string;
  maxPrice?: string;
  inStock?: string;
  page?: string;
};

const featuredSlugs = [
  "arc-task-lamp",
  "handwoven-market-basket",
  "compact-digital-camera",
  "waxed-utility-jacket",
  "field-bottle",
  "stoneware-dinner-set",
  "portable-radio",
  "canvas-tote",
];

function positiveInteger(value: string | undefined, fallback: number) {
  if (!value) return fallback;
  const parsed = Number(value);
  return Number.isSafeInteger(parsed) && parsed >= 0 ? parsed : Number.NaN;
}

export default async function Home({
  searchParams,
}: {
  searchParams: Promise<Search>;
}) {
  const filters = await searchParams;
  const minPrice = positiveInteger(filters.minPrice, 0);
  const maxPrice = filters.maxPrice
    ? positiveInteger(filters.maxPrice, 0)
    : undefined;
  const page = positiveInteger(filters.page, 1);
  const invalidFilters =
    Number.isNaN(minPrice) ||
    Number.isNaN(page) ||
    page < 1 ||
    (maxPrice !== undefined && (Number.isNaN(maxPrice) || maxPrice < minPrice));
  const hasFilters = Boolean(
    filters.q?.trim() ||
    filters.category ||
    filters.minPrice ||
    filters.maxPrice ||
    filters.inStock,
  );
  const pagedCatalog = hasFilters || page > 1;
  const query = new URLSearchParams({ pageSize: pagedCatalog ? "8" : "50" });
  if (filters.q?.trim()) query.set("q", filters.q.trim());
  if (filters.category) query.set("category", filters.category);
  if (!Number.isNaN(minPrice) && minPrice > 0)
    query.set("minPrice", String(minPrice));
  if (maxPrice !== undefined && !Number.isNaN(maxPrice))
    query.set("maxPrice", String(maxPrice));
  if (filters.inStock === "true") query.set("inStock", "true");
  if (!Number.isNaN(page) && page > 1) query.set("page", String(page));

  let catalog: ProductPage | null = null;
  let categories: Category[] = [];
  let loadError = false;
  if (!invalidFilters) {
    try {
      const products = await apiGet<ProductPage>(
        `/catalog/products?${query.toString()}`,
      );
      const categoryResponse = await apiGet<{ items: Category[] }>(
        "/categories",
      );
      catalog = products;
      categories = categoryResponse.items;
    } catch (error) {
      console.error("catalog load failed", {
        name: error instanceof Error ? error.name : "Unknown",
        message: error instanceof Error ? error.message : String(error),
        status: error instanceof APIError ? error.status : undefined,
        cause:
          error instanceof Error && error.cause instanceof Error
            ? error.cause.message
            : undefined,
      });
      loadError = true;
    }
  }
  const visibleProducts = catalog
    ? pagedCatalog
      ? catalog.items.slice(0, 8)
      : featuredSlugs
          .map((slug) => catalog.items.find((product) => product.slug === slug))
          .filter((product): product is ProductPage["items"][number] =>
            Boolean(product),
          )
    : [];
  const pages = catalog
    ? Math.max(1, Math.ceil(catalog.total / catalog.pageSize))
    : 1;
  const hasNextPage = Boolean(
    catalog && ((!pagedCatalog && catalog.total > 8) || catalog.page < pages),
  );
  const loadMoreQuery = new URLSearchParams(query);
  loadMoreQuery.delete("page");
  loadMoreQuery.delete("pageSize");
  const catalogKey = query.toString();
  const catalogReturnQuery = new URLSearchParams();
  if (filters.q?.trim()) catalogReturnQuery.set("q", filters.q.trim());
  if (filters.category) catalogReturnQuery.set("category", filters.category);
  if (filters.minPrice) catalogReturnQuery.set("minPrice", filters.minPrice);
  if (filters.maxPrice) catalogReturnQuery.set("maxPrice", filters.maxPrice);
  if (filters.inStock === "true") catalogReturnQuery.set("inStock", "true");
  if (filters.page) catalogReturnQuery.set("page", filters.page);
  const catalogReturnTo = `/${catalogReturnQuery.size ? `?${catalogReturnQuery}` : ""}#catalog`;
  const filterFormKey = JSON.stringify(filters);

  return (
    <main className="home-page">
      <section className="home-hero" aria-labelledby="home-title">
        <aside className="hero-kicker" aria-hidden="true">
          <span>Curated goods</span>
          <i />
          <b>01</b>
        </aside>
        <div className="hero-copy">
          <h1 id="home-title">
            Good things.
            <br />
            Still in
            <br />
            stock.
          </h1>
          <p>
            Independent stores.
            <br />
            Verified sellers.
            <br />
            Live inventory.
          </p>
        </div>
        <div className="hero-visual">
          <Image
            src="/images/home-hero-collage-v3.webp"
            alt="Task lamp, woven basket, camera, and waxed jacket"
            fill
            priority
            sizes="55vw"
          />
        </div>
        <aside className="hero-pages" aria-label="Featured collections">
          <span className="active">01</span>
          <span>XX</span>
          <span>XX</span>
        </aside>
      </section>
      <section
        id="catalog"
        className="catalog home-catalog"
        aria-labelledby="catalog-title"
      >
        <h2 id="catalog-title" className="sr-only">
          Catalog
        </h2>
        <form key={filterFormKey} className="filters" method="get" action="/">
          <div className="field search-field">
            <label className="search-icon" htmlFor="q" aria-label="Search">
              <svg aria-hidden="true" viewBox="0 0 24 24">
                <circle cx="10.5" cy="10.5" r="6.5" />
                <path d="m15.5 15.5 5 5" />
              </svg>
            </label>
            <input
              id="q"
              name="q"
              defaultValue={filters.q}
              maxLength={100}
              placeholder="Search products"
            />
          </div>
          <div className="field">
            <label className="sr-only" htmlFor="category">
              Category
            </label>
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
            <label className="sr-only" htmlFor="maxPrice">
              Price
            </label>
            <select
              id="maxPrice"
              name="maxPrice"
              defaultValue={filters.maxPrice ?? ""}
            >
              <option value="">All prices</option>
              <option value="500000">Up to IDR 500.000</option>
              <option value="1000000">Up to IDR 1.000.000</option>
              <option value="2500000">Up to IDR 2.500.000</option>
            </select>
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
            <button className="button button-primary" type="submit">
              Apply filters <span>→</span>
            </button>
            <CatalogClearLink className="button button-secondary">
              Clear
            </CatalogClearLink>
          </div>
        </form>
        {invalidFilters ? (
          <div className="state-panel" role="alert">
            <h2>Invalid filters</h2>
            <p>
              Use positive prices, minimum not above maximum, and page 1 or
              higher.
            </p>
            <CatalogClearLink className="button button-secondary">
              Clear filters
            </CatalogClearLink>
          </div>
        ) : loadError ? (
          <div className="state-panel" role="alert">
            <h2>Catalog unavailable</h2>
            <p>API could not be reached. Start local services, then reload.</p>
          </div>
        ) : visibleProducts.length > 0 && catalog ? (
          <CatalogProductList
            initialProducts={visibleProducts}
            initialPosition={(pagedCatalog ? page - 1 : 0) * 8}
            initialHasMore={hasNextPage}
            filterQuery={loadMoreQuery.toString()}
            catalogKey={catalogKey}
            returnTo={catalogReturnTo}
            loadPage={loadCatalogPage}
          />
        ) : (
          <div className="state-panel">
            <h2>No products match</h2>
            <p>Clear filters or use another search term.</p>
            <CatalogClearLink className="button button-secondary">
              Reset catalog
            </CatalogClearLink>
          </div>
        )}
      </section>
    </main>
  );
}
