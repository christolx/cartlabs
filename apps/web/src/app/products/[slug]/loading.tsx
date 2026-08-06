import { SiteHeader } from "@/components/site-header";

export default function LoadingProduct() {
  return <><SiteHeader /><main className="product-page shell"><div className="detail-skeleton" aria-label="Loading product" /></main></>;
}
