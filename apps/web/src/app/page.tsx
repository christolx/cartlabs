import Link from "next/link";
import { ProductCard } from "@/components/product-card";
import { apiGet, type Category, type ProductPage } from "@/lib/api/client";

type Search = { q?: string; category?: string; minPrice?: string; maxPrice?: string; inStock?: string; page?: string };

function positiveInteger(value: string | undefined, fallback: number) {
  if (!value) return fallback;
  const parsed = Number(value);
  return Number.isSafeInteger(parsed) && parsed >= 0 ? parsed : Number.NaN;
}

function pageHref(filters: Search, page: number) {
  const query = new URLSearchParams();
  for (const [key, value] of Object.entries(filters)) if (value && key !== "page") query.set(key, value);
  query.set("page", String(page));
  return `/?${query.toString()}#catalog`;
}

export default async function Home({ searchParams }: { searchParams: Promise<Search> }) {
  const filters = await searchParams;
  const minPrice = positiveInteger(filters.minPrice, 0);
  const maxPrice = filters.maxPrice ? positiveInteger(filters.maxPrice, 0) : undefined;
  const page = positiveInteger(filters.page, 1);
  const invalidFilters = Number.isNaN(minPrice) || Number.isNaN(page) || page < 1 || (maxPrice !== undefined && (Number.isNaN(maxPrice) || maxPrice < minPrice));
  const query = new URLSearchParams({ pageSize: "20" });
  if (filters.q?.trim()) query.set("q", filters.q.trim());
  if (filters.category) query.set("category", filters.category);
  if (!Number.isNaN(minPrice) && minPrice > 0) query.set("minPrice", String(minPrice));
  if (maxPrice !== undefined && !Number.isNaN(maxPrice)) query.set("maxPrice", String(maxPrice));
  if (filters.inStock === "true") query.set("inStock", "true");
  if (!Number.isNaN(page) && page > 1) query.set("page", String(page));

  let catalog: ProductPage | null = null;
  let categories: Category[] = [];
  let loadError = false;
  if (!invalidFilters) {
    try {
      const [products, categoryResponse] = await Promise.all([
        apiGet<ProductPage>(`/catalog/products?${query.toString()}`),
        apiGet<{ items: Category[] }>("/categories"),
      ]);
      catalog = products;
      categories = categoryResponse.items;
    } catch {
      loadError = true;
    }
  }
  const pages = catalog ? Math.max(1, Math.ceil(catalog.total / catalog.pageSize)) : 1;

  return <main>
    <section className="catalog-intro shell" aria-labelledby="home-title"><div><p className="eyebrow">Verified independent stores</p><h1 id="home-title">Useful goods, live inventory.</h1></div><p>Discover approved products from independent sellers. Store moderation, stock, checkout, and fulfillment use real marketplace state.</p></section>
    <section id="catalog" className="catalog shell" aria-labelledby="catalog-title">
      <div className="section-heading"><h2 id="catalog-title">Catalog</h2><p>{catalog ? `${catalog.total} published products` : "Live inventory from approved stores"}</p></div>
      <form className="filters" method="get" action="/">
        <div className="field search-field"><label htmlFor="q">Search products</label><input id="q" name="q" defaultValue={filters.q} maxLength={100} placeholder="Basket, lamp, textile" /></div>
        <div className="field"><label htmlFor="category">Category</label><select id="category" name="category" defaultValue={filters.category ?? ""}><option value="">All categories</option>{categories.map((category) => <option key={category.id} value={category.slug}>{category.name}</option>)}</select></div>
        <div className="field"><label htmlFor="minPrice">Minimum price (IDR)</label><input id="minPrice" name="minPrice" type="number" min="0" defaultValue={filters.minPrice} /></div>
        <div className="field"><label htmlFor="maxPrice">Maximum price (IDR)</label><input id="maxPrice" name="maxPrice" type="number" min="0" defaultValue={filters.maxPrice} /></div>
        <label className="check-field"><input type="checkbox" name="inStock" value="true" defaultChecked={filters.inStock === "true"} />In stock only</label>
        <div className="filter-actions"><button className="button button-primary" type="submit">Apply filters</button><Link className="button button-secondary" href="/#catalog">Clear</Link></div>
      </form>
      {invalidFilters ? <div className="state-panel" role="alert"><h2>Invalid filters</h2><p>Use positive prices, minimum not above maximum, and page 1 or higher.</p><Link className="button button-secondary" href="/#catalog">Clear filters</Link></div> : loadError ? <div className="state-panel" role="alert"><h2>Catalog unavailable</h2><p>API could not be reached. Start local services, then reload.</p></div> : catalog && catalog.items.length > 0 ? <><div className="product-grid">{catalog.items.map((product) => <ProductCard key={product.id} product={product} />)}</div><nav className="pagination" aria-label="Catalog pages"><Link className="button button-secondary" aria-disabled={catalog.page <= 1} href={catalog.page <= 1 ? pageHref(filters, 1) : pageHref(filters, catalog.page - 1)}>Previous</Link><span>Page {catalog.page} of {pages}</span><Link className="button button-secondary" aria-disabled={catalog.page >= pages} href={catalog.page >= pages ? pageHref(filters, pages) : pageHref(filters, catalog.page + 1)}>Next</Link></nav></> : <div className="state-panel"><h2>No products match</h2><p>Clear filters or use another search term.</p><Link className="button button-secondary" href="/#catalog">Reset catalog</Link></div>}
    </section>
  </main>;
}
