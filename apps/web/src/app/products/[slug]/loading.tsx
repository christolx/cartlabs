export default function LoadingProduct() {
  return (
    <main className="product-page shell" aria-busy="true">
      <span className="sr-only">Loading product</span>
      <div className="product-loading-grid" aria-hidden="true">
        <div className="detail-skeleton product-loading-media" />
        <div className="product-loading-copy">
          <div className="skeleton-line short" />
          <div className="skeleton-line product-loading-title" />
          <div className="skeleton-line" />
          <div className="product-loading-facts" />
          <div className="product-loading-action" />
        </div>
      </div>
    </main>
  );
}
