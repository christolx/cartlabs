"use server";

import { apiGet, type ProductPage } from "@/lib/api/client";

function nonNegativeInteger(value: string | null, name: string) {
  if (value === null) return undefined;
  const parsed = Number(value);
  if (!Number.isSafeInteger(parsed) || parsed < 0)
    throw new Error(`Invalid ${name}`);
  return parsed;
}

export async function loadCatalogPage(
  filterQuery: string,
  excludeIds: string[],
) {
  if (
    !Array.isArray(excludeIds) ||
    excludeIds.length > 1_000 ||
    excludeIds.some((id) => !/^[0-9a-f-]{36}$/.test(id))
  ) {
    throw new Error("Invalid excluded products");
  }
  const filters = new URLSearchParams(filterQuery);
  const query = new URLSearchParams({ pageSize: "50" });
  const search = filters.get("q")?.trim();
  const category = filters.get("category");
  const minPrice = nonNegativeInteger(filters.get("minPrice"), "minimum price");
  const maxPrice = nonNegativeInteger(filters.get("maxPrice"), "maximum price");

  if (search) {
    if (search.length > 100) throw new Error("Invalid search");
    query.set("q", search);
  }
  if (category) {
    if (category.length > 100 || !/^[a-z0-9]+(?:-[a-z0-9]+)*$/.test(category))
      throw new Error("Invalid category");
    query.set("category", category);
  }
  if (minPrice !== undefined) query.set("minPrice", String(minPrice));
  if (maxPrice !== undefined) query.set("maxPrice", String(maxPrice));
  if (minPrice !== undefined && maxPrice !== undefined && minPrice > maxPrice)
    throw new Error("Invalid price range");
  if (filters.get("inStock") === "true") query.set("inStock", "true");

  const excluded = new Set(excludeIds);
  const items: ProductPage["items"] = [];
  let page = 1;
  let hasAnotherPage = false;

  do {
    query.set("page", String(page));
    const result = await apiGet<ProductPage>(
      `/catalog/products?${query.toString()}`,
    );
    for (const product of result.items) {
      if (!excluded.has(product.id)) items.push(product);
      if (items.length === 9) break;
    }
    hasAnotherPage = result.page * result.pageSize < result.total;
    page += 1;
  } while (items.length < 9 && hasAnotherPage);

  return {
    items: items.slice(0, 8),
    hasMore: items.length > 8 || hasAnotherPage,
  };
}
