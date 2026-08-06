import Image from "next/image";
import Link from "next/link";
import { ProductCard } from "@/components/product-card";
import { SiteHeader } from "@/components/site-header";
import { apiGet, type Category, type ProductPage } from "@/lib/api/client";

type Search = { q?: string; category?: string; inStock?: string };

export default async function Home({ searchParams }: { searchParams: Promise<Search> }) {
  const filters = await searchParams;
  const query = new URLSearchParams();
  if (filters.q) query.set("q", filters.q);
  if (filters.category) query.set("category", filters.category);
  if (filters.inStock === "true") query.set("inStock", "true");

  let catalog: ProductPage | null = null;
  let categories: Category[] = [];
  let loadError = false;
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

  return (
    <>
      <SiteHeader />
      <main>
        <section className="hero shell" aria-labelledby="hero-title">
          <Image
            className="hero-image"
            src="/images/marketplace-hero-stock.webp"
            alt="Flower, produce, and woven-goods stalls inside a covered market"
            fill
            priority
            sizes="(max-width: 767px) 100vw, 1400px"
          />
          <div className="hero-scrim" />
          <div className="hero-copy">
            <p className="eyebrow">Independent sellers, one market</p>
            <h1 id="hero-title">Good objects. Clear provenance.</h1>
            <p>Browse practical goods from verified independent stores across Indonesia.</p>
            <div className="hero-actions">
              <Link className="button button-primary" href="#catalog">Browse goods</Link>
              <Link className="button button-secondary" href="/demo">Try demo</Link>
            </div>
          </div>
        </section>

        <section id="catalog" className="catalog shell" aria-labelledby="catalog-title">
          <div className="section-heading">
            <h2 id="catalog-title">Marketplace catalog</h2>
            <p>{catalog ? `${catalog.total} published products` : "Live inventory from verified stores"}</p>
          </div>
          <form className="filters" method="get" action="/">
            <div className="field search-field">
              <label htmlFor="q">Search products</label>
              <input id="q" name="q" defaultValue={filters.q} placeholder="Basket, lamp, textile" />
            </div>
            <div className="field">
              <label htmlFor="category">Category</label>
              <select id="category" name="category" defaultValue={filters.category ?? ""}>
                <option value="">All categories</option>
                {categories.map((category) => <option key={category.id} value={category.slug}>{category.name}</option>)}
              </select>
            </div>
            <label className="check-field">
              <input type="checkbox" name="inStock" value="true" defaultChecked={filters.inStock === "true"} />
              In stock only
            </label>
            <button className="button button-primary filter-button" type="submit">Apply filters</button>
          </form>

          {loadError ? (
            <div className="state-panel" role="alert">
              <h3>Catalog unavailable</h3>
              <p>API could not be reached. Start local services, then reload this page.</p>
            </div>
          ) : catalog && catalog.items.length > 0 ? (
            <div className="product-grid">
              {catalog.items.map((product) => <ProductCard key={product.id} product={product} />)}
            </div>
          ) : (
            <div className="state-panel">
              <h3>No products match</h3>
              <p>Clear a filter or use another search term.</p>
              <Link className="text-link" href="/">Reset catalog</Link>
            </div>
          )}
        </section>
      </main>
      <footer className="site-footer">
        <div className="shell footer-inner">
          <p>Cartlabs demonstrates marketplace engineering through working vertical slices.</p>
          <Link href="/demo">Open role demo</Link>
        </div>
      </footer>
    </>
  );
}
