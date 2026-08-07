"use client";

import Image from "next/image";
import { useCallback, useState, type FormEvent } from "react";
import type { components } from "@/lib/api/schema";

type Role = components["schemas"]["Role"];
type Session = components["schemas"]["Session"];
type Store = components["schemas"]["Store"];
type Product = components["schemas"]["ProductDetail"];
type ProductPage = components["schemas"]["ProductPage"];
type Category = components["schemas"]["Category"];
type Cart = components["schemas"]["Cart"];
type Purchase = components["schemas"]["Purchase"];
type SellerOrder = components["schemas"]["SellerOrder"];
type Notification = components["schemas"]["Notification"];
type AdminOverview = components["schemas"]["AdminOverview"];
type AuditEvent = components["schemas"]["AuditEvent"];
type Review = components["schemas"]["Review"];

const DEMO_PRODUCT_IMAGE_URL = "/images/shared-product.webp";

type Workspace = {
  store?: Store;
  stores?: Store[];
  products: Product[];
  categories: Category[];
  cart?: Cart;
  purchases?: Purchase[];
  orders?: SellerOrder[];
  notifications?: Notification[];
  overview?: AdminOverview;
  auditEvents?: AuditEvent[];
};

async function request<T>(path: string, token: string, init?: RequestInit): Promise<T> {
  const headers = new Headers(init?.headers);
  headers.set("Accept", "application/json");
  if (token) headers.set("Authorization", `Bearer ${token}`);
  if (init?.body) headers.set("Content-Type", "application/json");
  const response = await fetch(`/api/backend${path}`, { ...init, headers });
  if (!response.ok) {
    const problem = (await response.json().catch(() => ({}))) as { detail?: string };
    throw new Error(problem.detail ?? `Request failed with status ${response.status}`);
  }
  if (response.status === 204) return undefined as T;
  return (await response.json()) as T;
}

function formatMoney(value: number, currency: string) {
  return new Intl.NumberFormat("id-ID", {
    style: "currency",
    currency,
    maximumFractionDigits: 0,
  }).format(value);
}

export function DemoConsole() {
  const [session, setSession] = useState<Session | null>(null);
  const [workspace, setWorkspace] = useState<Workspace>({ products: [], categories: [] });
  const [busy, setBusy] = useState(false);
  const [message, setMessage] = useState("");

  const loadWorkspace = useCallback(async (current: Session) => {
    const token = current.accessToken;
    if (current.user.role === "seller") {
      const [storeResponse, products, categories, orders, notifications] = await Promise.all([
        request<Store>("/seller/store", token).catch((error: Error) =>
          error.message === "resource not found" ? undefined : Promise.reject(error),
        ),
        request<{ items: Product[] }>("/seller/products", token),
        request<{ items: Category[] }>("/categories", token),
        request<{ items: SellerOrder[] }>("/seller/orders", token),
        request<{ items: Notification[] }>("/notifications", token),
      ]);
      setWorkspace({
        store: storeResponse,
        products: products.items,
        categories: categories.items,
        orders: orders.items,
        notifications: notifications.items,
      });
      return;
    }
    if (current.user.role === "admin") {
      const [stores, products, overview, auditEvents] = await Promise.all([
        request<{ items: Store[] }>("/admin/stores", token),
        request<{ items: Product[] }>("/admin/products", token),
        request<AdminOverview>("/admin/overview", token),
        request<{ items: AuditEvent[] }>("/admin/audit-events", token),
      ]);
      setWorkspace({ stores: stores.items, products: products.items, categories: [], overview, auditEvents: auditEvents.items });
      return;
    }

    const [catalog, cart, purchases, notifications] = await Promise.all([
      request<ProductPage>("/catalog/products?pageSize=24", token),
      request<Cart>("/cart", token),
      request<{ items: Purchase[] }>("/purchases", token),
      request<{ items: Notification[] }>("/notifications", token),
    ]);
    const products = await Promise.all(
      catalog.items.map((product) =>
        request<Product>(`/catalog/products/${encodeURIComponent(product.slug)}`, token),
      ),
    );
    setWorkspace({
      products,
      categories: [],
      cart,
      purchases: purchases.items,
      notifications: notifications.items,
    });
  }, []);

  async function runAction(action: () => Promise<void>, fallback: string) {
    setBusy(true);
    setMessage("");
    try {
      await action();
    } catch (error) {
      setMessage(error instanceof Error ? error.message : fallback);
    } finally {
      setBusy(false);
    }
  }

  async function login(role: Role) {
    await runAction(async () => {
      const current = await request<Session>("/auth/demo-login", "", {
        method: "POST",
        body: JSON.stringify({ role }),
      });
      setSession(current);
      await loadWorkspace(current);
      setMessage(`Signed in as ${current.user.displayName}.`);
    }, "Login failed");
  }

  async function refreshWorkspace() {
    if (!session) return;
    await runAction(async () => {
      await loadWorkspace(session);
      setMessage("Workspace refreshed.");
    }, "Refresh failed");
  }

  async function submitStore(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!session) return;
    const form = event.currentTarget;
    const values = Object.fromEntries(new FormData(form));
    await runAction(async () => {
      await request<Store>("/seller/store", session.accessToken, {
        method: "POST",
        body: JSON.stringify(values),
      });
      await loadWorkspace(session);
      setMessage("Store submitted for admin review.");
      form.reset();
    }, "Store creation failed");
  }

  async function submitProduct(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!session) return;
    const form = event.currentTarget;
    const data = new FormData(form);
    await runAction(async () => {
      const product = await request<Product>("/seller/products", session.accessToken, {
        method: "POST",
        body: JSON.stringify({
          categoryId: data.get("categoryId"),
          name: data.get("name"),
          slug: data.get("slug"),
          description: data.get("description"),
        }),
      });
      await request(`/seller/products/${product.id}/variants`, session.accessToken, {
        method: "POST",
        body: JSON.stringify({
          sku: data.get("sku"),
          name: data.get("variantName"),
          attributes: {},
          priceMinor: Number(data.get("priceMinor")),
          currency: "IDR",
          stock: Number(data.get("stock")),
        }),
      });
      await request(`/seller/products/${product.id}/images`, session.accessToken, {
        method: "POST",
        body: JSON.stringify({
          url: DEMO_PRODUCT_IMAGE_URL,
          altText: data.get("name"),
          position: 0,
        }),
      });
      await request(`/seller/products/${product.id}/publish`, session.accessToken, { method: "POST" });
      await loadWorkspace(session);
      setMessage("Product published and waiting for admin review.");
      form.reset();
    }, "Product creation failed");
  }

  async function moderate(kind: "stores" | "products", id: string, status: "approved" | "rejected") {
    if (!session) return;
    await runAction(async () => {
      await request(`/admin/${kind}/${id}/moderation`, session.accessToken, {
        method: "PATCH",
        body: JSON.stringify({ status, note: "Reviewed in web demo" }),
      });
      await loadWorkspace(session);
      setMessage(`${kind === "stores" ? "Store" : "Product"} ${status}.`);
    }, "Moderation failed");
  }

  async function setCartQuantity(variantId: string, quantity: number) {
    if (!session) return;
    await runAction(async () => {
      const cart = await request<Cart>(`/cart/items/${variantId}`, session.accessToken, quantity > 0
        ? { method: "PUT", body: JSON.stringify({ quantity }) }
        : { method: "DELETE" });
      setWorkspace((current) => ({ ...current, cart }));
      setMessage(quantity > 0 ? "Cart updated." : "Item removed.");
    }, "Cart update failed");
  }

  async function checkout() {
    if (!session) return;
    await runAction(async () => {
      const purchase = await request<Purchase>("/checkout", session.accessToken, {
        method: "POST",
        headers: { "Idempotency-Key": crypto.randomUUID() },
      });
      await loadWorkspace(session);
      setMessage(`Purchase ${purchase.reference} reserved for 15 minutes.`);
    }, "Checkout failed");
  }

  async function completePayment(purchaseId: string, outcome: "succeeded" | "failed") {
    if (!session) return;
    await runAction(async () => {
      const purchase = await request<Purchase>(`/purchases/${purchaseId}/pay`, session.accessToken, {
        method: "POST",
        body: JSON.stringify({ outcome }),
      });
      await loadWorkspace(session);
      setMessage(`Payment ${purchase.paymentStatus} for ${purchase.reference}.`);
    }, "Payment simulation failed");
  }

  async function cancelPurchase(purchaseId: string) {
    if (!session) return;
    await runAction(async () => {
      const purchase = await request<Purchase>(`/purchases/${purchaseId}/cancel`, session.accessToken, {
        method: "POST",
        body: JSON.stringify({ reason: "Cancelled from buyer demo" }),
      });
      await loadWorkspace(session);
      setMessage(`Purchase ${purchase.reference} cancelled and inventory restored.`);
    }, "Purchase cancellation failed");
  }

  async function createReview(purchaseItemId: string, productName: string) {
    if (!session) return;
    await runAction(async () => {
      await request<Review>("/reviews", session.accessToken, {
        method: "POST",
        body: JSON.stringify({ purchaseItemId, rating: 5, title: "Verified delivery", body: `${productName} arrived as expected.` }),
      });
      setMessage(`Review published for ${productName}.`);
    }, "Review creation failed");
  }

  async function updateSellerOrder(orderId: string, status: "processing" | "shipped" | "delivered" | "cancelled") {
    if (!session) return;
    await runAction(async () => {
      await request<SellerOrder>(`/seller/orders/${orderId}/status`, session.accessToken, {
        method: "PATCH",
        body: JSON.stringify({ status, reason: status === "cancelled" ? "Cancelled from seller demo" : "" }),
      });
      await loadWorkspace(session);
      setMessage(`Seller order moved to ${status}.`);
    }, "Fulfillment update failed");
  }

  return (
    <div className="demo-console">
      <div className="console-toolbar">
        <div className="role-switcher" role="group" aria-label="Demo role selection">
          {(["buyer", "seller", "admin"] as Role[]).map((role) => (
            <button
              key={role}
              className={session?.user.role === role ? "role-active" : ""}
              aria-pressed={session?.user.role === role}
              disabled={busy}
              onClick={() => login(role)}
            >
              {role}
            </button>
          ))}
        </div>
        {session && <button className="quiet-button" disabled={busy} onClick={refreshWorkspace}>Refresh</button>}
      </div>
      <p className="demo-message" aria-live="polite">
        {busy ? "Working..." : message || "Choose a role. Demo mode must be enabled in API configuration."}
      </p>

      {session?.user.role === "buyer" && (
        <BuyerWorkspace
          workspace={workspace}
          busy={busy}
          setCartQuantity={setCartQuantity}
          checkout={checkout}
          completePayment={completePayment}
          cancelPurchase={cancelPurchase}
          createReview={createReview}
        />
      )}

      {session?.user.role === "seller" && (
        <div className="workspace-grid">
          <section className="workspace-panel">
            <p className="panel-kicker">Storefront</p>
            <h2>Seller store</h2>
            {workspace.store ? (
              <>
                <p className="status-line">
                  <strong>{workspace.store.name}</strong>
                  <span data-status={workspace.store.status}>{workspace.store.status}</span>
                </p>
                <p>{workspace.store.description}</p>
              </>
            ) : (
              <form className="stack-form" onSubmit={submitStore}>
                <Field label="Store name" name="name" minLength={2} required />
                <Field label="Store slug" name="slug" pattern="[a-z0-9]+(?:-[a-z0-9]+)*" required />
                <Field label="Description" name="description" required />
                <button className="button button-primary" disabled={busy}>Submit store</button>
              </form>
            )}
          </section>
          <section className="workspace-panel">
            <p className="panel-kicker">Catalog</p>
            <h2>Publish product</h2>
            <p>Requires approved store. Product enters admin review after publish.</p>
            <form className="stack-form two-column-form" onSubmit={submitProduct}>
              <Field label="Product name" name="name" required />
              <Field label="Product slug" name="slug" pattern="[a-z0-9]+(?:-[a-z0-9]+)*" required />
              <label>
                <span>Category</span>
                <select name="categoryId" required>
                  {workspace.categories.map((item) => <option key={item.id} value={item.id}>{item.name}</option>)}
                </select>
              </label>
              <Field label="SKU" name="sku" required />
              <Field label="Variant name" name="variantName" required />
              <Field label="Price (IDR)" name="priceMinor" type="number" min={0} required />
              <Field label="Starting stock" name="stock" type="number" min={1} required />
              <label><span>Product image</span><input value="Shared demo image" readOnly /></label>
              <label className="form-wide"><span>Description</span><textarea name="description" rows={3} required /></label>
              <button className="button button-primary form-wide" disabled={busy}>Create and publish</button>
            </form>
          </section>
          <section className="workspace-panel workspace-wide">
            <p className="panel-kicker">Inventory</p>
            <h2>Owned products</h2>
            <ProductRows products={workspace.products} />
          </section>
          <section className="workspace-panel workspace-wide">
            <p className="panel-kicker">Fulfillment</p>
            <h2>Seller orders</h2>
            <SellerOrderRows orders={workspace.orders ?? []} busy={busy} updateOrder={updateSellerOrder} />
          </section>
          <section className="workspace-panel workspace-wide">
            <p className="panel-kicker">Events</p>
            <h2>Notifications</h2>
            <NotificationRows notifications={workspace.notifications ?? []} />
          </section>
        </div>
      )}

      {session?.user.role === "admin" && (
        <div className="workspace-grid">
          <section className="workspace-panel workspace-wide">
            <p className="panel-kicker">Live totals</p>
            <h2>Marketplace overview</h2>
            <OverviewGrid overview={workspace.overview} />
          </section>
          <section className="workspace-panel">
            <p className="panel-kicker">Review queue</p>
            <h2>Store moderation</h2>
            <ModerationRows items={workspace.stores ?? []} kind="stores" busy={busy} moderate={moderate} />
          </section>
          <section className="workspace-panel">
            <p className="panel-kicker">Review queue</p>
            <h2>Product moderation</h2>
            <ModerationRows items={workspace.products} kind="products" busy={busy} moderate={moderate} />
          </section>
          <section className="workspace-panel workspace-wide">
            <p className="panel-kicker">Immutable activity</p>
            <h2>Audit trail</h2>
            <AuditRows events={workspace.auditEvents ?? []} />
          </section>
        </div>
      )}
    </div>
  );
}

function BuyerWorkspace({
  workspace,
  busy,
  setCartQuantity,
  checkout,
  completePayment,
  cancelPurchase,
  createReview,
}: {
  workspace: Workspace;
  busy: boolean;
  setCartQuantity: (variantId: string, quantity: number) => void;
  checkout: () => void;
  completePayment: (purchaseId: string, outcome: "succeeded" | "failed") => void;
  cancelPurchase: (purchaseId: string) => void;
  createReview: (purchaseItemId: string, productName: string) => void;
}) {
  const cartQuantities = new Map(
    workspace.cart?.stores.flatMap((store) => store.items.map((item) => [item.variantId, item.quantity] as const)),
  );

  return (
    <div className="buyer-workspace">
      <section className="workspace-panel buyer-market">
        <div className="panel-heading-row">
          <div>
            <p className="panel-kicker">Approved catalog</p>
            <h2>Build a two-seller cart</h2>
          </div>
          <span className="count-badge">{workspace.products.length} products</span>
        </div>
        <div className="buyer-product-grid">
          {workspace.products.map((product) => {
            const variant = product.variants.find((item) => item.active && item.stock > 0);
            const quantity = variant ? cartQuantities.get(variant.id) ?? 0 : 0;
            return (
              <article className="buyer-product" key={product.id}>
                <div className="buyer-product-image">
                  <Image
                    src={product.images[0]?.url || DEMO_PRODUCT_IMAGE_URL}
                    alt={product.images[0]?.altText || product.name}
                    fill
                    sizes="(max-width: 767px) 100vw, 360px"
                  />
                </div>
                <div className="buyer-product-copy">
                  <p>{product.storeName}</p>
                  <h3>{product.name}</h3>
                  {variant ? (
                    <>
                      <div className="product-price-line">
                        <strong>{formatMoney(variant.priceMinor, variant.currency)}</strong>
                        <span>{variant.stock} available</span>
                      </div>
                      <button
                        className="button button-primary"
                        disabled={busy || quantity >= variant.stock}
                        onClick={() => setCartQuantity(variant.id, quantity + 1)}
                      >
                        {quantity ? `Add another · ${quantity} in cart` : "Add to cart"}
                      </button>
                    </>
                  ) : <p className="empty-copy">Out of stock</p>}
                </div>
              </article>
            );
          })}
        </div>
      </section>

      <section className="workspace-panel cart-panel">
        <div className="panel-heading-row">
          <div>
            <p className="panel-kicker">Grouped by seller</p>
            <h2>Cart</h2>
          </div>
          <span className="count-badge">{workspace.cart?.totalQuantity ?? 0} items</span>
        </div>
        <CartContents cart={workspace.cart} busy={busy} setCartQuantity={setCartQuantity} />
        {!!workspace.cart?.totalQuantity && (
          <div className="cart-total">
            <div><span>Total</span><strong>{formatMoney(workspace.cart.subtotalMinor, workspace.cart.currency)}</strong></div>
            <button className="button button-primary" disabled={busy} onClick={checkout}>Reserve and checkout</button>
          </div>
        )}
      </section>

      <section className="workspace-panel">
        <p className="panel-kicker">Immutable snapshots</p>
        <h2>Purchases</h2>
        <PurchaseRows purchases={workspace.purchases ?? []} busy={busy} completePayment={completePayment} cancelPurchase={cancelPurchase} createReview={createReview} />
      </section>

      <section className="workspace-panel">
        <p className="panel-kicker">Outbox delivery</p>
        <h2>Notifications</h2>
        <NotificationRows notifications={workspace.notifications ?? []} />
      </section>
    </div>
  );
}

function CartContents({
  cart,
  busy,
  setCartQuantity,
}: {
  cart?: Cart;
  busy: boolean;
  setCartQuantity: (variantId: string, quantity: number) => void;
}) {
  if (!cart?.stores.length) return <p className="empty-copy">Cart empty. Add products from both sellers.</p>;
  return (
    <div className="cart-store-list">
      {cart.stores.map((store) => (
        <div className="cart-store" key={store.storeId}>
          <div className="cart-store-heading">
            <strong>{store.storeName}</strong>
            <span>{formatMoney(store.subtotalMinor, cart.currency)}</span>
          </div>
          {store.items.map((item) => (
            <div className="cart-item" key={item.variantId}>
              <Image src={item.imageUrl || DEMO_PRODUCT_IMAGE_URL} alt={item.productName} width={72} height={54} />
              <div>
                <strong>{item.productName}</strong>
                <span>{item.variantName} · {formatMoney(item.unitPriceMinor, item.currency)}</span>
              </div>
              <div className="quantity-control" role="group" aria-label={`Quantity for ${item.productName}`}>
                <button disabled={busy} aria-label="Decrease quantity" onClick={() => setCartQuantity(item.variantId, item.quantity - 1)}>−</button>
                <span>{item.quantity}</span>
                <button disabled={busy || item.quantity >= item.availableStock} aria-label="Increase quantity" onClick={() => setCartQuantity(item.variantId, item.quantity + 1)}>+</button>
              </div>
            </div>
          ))}
        </div>
      ))}
    </div>
  );
}

function PurchaseRows({
  purchases,
  busy,
  completePayment,
  cancelPurchase,
  createReview,
}: {
  purchases: Purchase[];
  busy: boolean;
  completePayment: (purchaseId: string, outcome: "succeeded" | "failed") => void;
  cancelPurchase: (purchaseId: string) => void;
  createReview: (purchaseItemId: string, productName: string) => void;
}) {
  if (!purchases.length) return <p className="empty-copy">No purchases yet.</p>;
  return (
    <div className="purchase-list">
      {purchases.map((purchase) => (
        <article className="purchase-record" key={purchase.id}>
          <div className="purchase-heading">
            <div><strong>{purchase.reference}</strong><span>{purchase.sellerOrders.length} seller orders</span></div>
            <span data-status={purchase.status}>{purchase.status.replaceAll("_", " ")}</span>
          </div>
          <div className="purchase-summary">
            <strong>{formatMoney(purchase.totalMinor, purchase.currency)}</strong>
            <span>Reserved until {new Date(purchase.reservationExpiresAt).toLocaleTimeString("id-ID", { hour: "2-digit", minute: "2-digit" })}</span>
          </div>
          <SellerOrderRows orders={purchase.sellerOrders} compact />
          {purchase.status === "pending_payment" && (
            <div className="payment-actions">
              <button className="button button-primary" disabled={busy} onClick={() => completePayment(purchase.id, "succeeded")}>Complete payment</button>
              <button className="quiet-button danger-button" disabled={busy} onClick={() => completePayment(purchase.id, "failed")}>Simulate failure</button>
            </div>
          )}
          {(purchase.status === "pending_payment" || purchase.status === "paid") && purchase.sellerOrders.every((order) => order.status === "pending_payment" || order.status === "paid") && (
            <button className="quiet-button danger-button purchase-cancel" disabled={busy} onClick={() => cancelPurchase(purchase.id)}>Cancel purchase</button>
          )}
          {purchase.sellerOrders.filter((order) => order.status === "delivered").flatMap((order) => order.items).map((item) => (
            <button className="quiet-button review-button" key={item.id} disabled={busy} onClick={() => createReview(item.id, item.productName)}>Review {item.productName}</button>
          ))}
        </article>
      ))}
    </div>
  );
}

function Field({ label, ...input }: { label: string } & React.InputHTMLAttributes<HTMLInputElement>) {
  return <label><span>{label}</span><input {...input} /></label>;
}

function ProductRows({ products }: { products: Product[] }) {
  if (!products.length) return <p className="empty-copy">No owned products yet.</p>;
  return (
    <div className="record-list">
      {products.map((product) => (
        <div className="record-row" key={product.id}>
          <div><strong>{product.name}</strong><span>{product.variants.length} variants</span></div>
          <span data-status={product.moderationStatus}>{product.status}, {product.moderationStatus}</span>
        </div>
      ))}
    </div>
  );
}

function SellerOrderRows({ orders, compact = false, busy = false, updateOrder }: { orders: SellerOrder[]; compact?: boolean; busy?: boolean; updateOrder?: (orderId: string, status: "processing" | "shipped" | "delivered" | "cancelled") => void }) {
  if (!orders.length) return <p className="empty-copy">No seller orders yet.</p>;
  return (
    <div className={compact ? "seller-order-list compact-orders" : "seller-order-list"}>
      {orders.map((order) => (
        <div className="seller-order" key={order.id}>
          <div>
            <strong>{order.storeName}</strong>
            <span>{order.reference} · {order.items.length} lines</span>
          </div>
          <div>
            <strong>{formatMoney(order.subtotalMinor, order.currency)}</strong>
            <span data-status={order.status}>{order.status.replaceAll("_", " ")}</span>
          </div>
          {updateOrder && <FulfillmentActions order={order} busy={busy} updateOrder={updateOrder} />}
        </div>
      ))}
    </div>
  );
}

function FulfillmentActions({ order, busy, updateOrder }: { order: SellerOrder; busy: boolean; updateOrder: (orderId: string, status: "processing" | "shipped" | "delivered" | "cancelled") => void }) {
  const next = order.status === "paid" ? "processing" : order.status === "processing" ? "shipped" : order.status === "shipped" ? "delivered" : null;
  if (!next && order.status !== "paid" && order.status !== "processing") return null;
  return (
    <div className="fulfillment-actions">
      {next && <button className="quiet-button" disabled={busy} onClick={() => updateOrder(order.id, next)}>Mark {next}</button>}
      {(order.status === "paid" || order.status === "processing") && <button className="quiet-button danger-button" disabled={busy} onClick={() => updateOrder(order.id, "cancelled")}>Cancel</button>}
    </div>
  );
}

function OverviewGrid({ overview }: { overview?: AdminOverview }) {
  if (!overview) return <p className="empty-copy">Overview unavailable.</p>;
  const values = [
    ["Users", overview.users], ["Approved stores", overview.approvedStores], ["Published products", overview.publishedProducts],
    ["Purchases", overview.purchases], ["Active orders", overview.activeSellerOrders], ["Delivered", overview.deliveredOrders],
  ] as const;
  return <div className="overview-grid">{values.map(([label, value]) => <div key={label}><span>{label}</span><strong>{value}</strong></div>)}<div className="overview-gmv"><span>Active GMV</span><strong>{formatMoney(overview.grossMerchandiseMinor, overview.currency)}</strong></div></div>;
}

function AuditRows({ events }: { events: AuditEvent[] }) {
  if (!events.length) return <p className="empty-copy">No audited activity yet.</p>;
  return <div className="audit-list">{events.map((event) => <div key={event.id}><div><strong>{event.action.replaceAll(".", " / ")}</strong><span>{event.actorName || event.actorRole}</span></div><time dateTime={event.createdAt}>{new Date(event.createdAt).toLocaleString("id-ID")}</time></div>)}</div>;
}

function NotificationRows({ notifications }: { notifications: Notification[] }) {
  if (!notifications.length) return <p className="empty-copy">No notifications yet.</p>;
  return (
    <div className="notification-list">
      {notifications.map((notification) => (
        <article key={notification.id}>
          <span>{notification.kind.replaceAll(".", " / ")}</span>
          <strong>{notification.title}</strong>
          <p>{notification.body}</p>
        </article>
      ))}
    </div>
  );
}

function ModerationRows({
  items,
  kind,
  busy,
  moderate,
}: {
  items: (Store | Product)[];
  kind: "stores" | "products";
  busy: boolean;
  moderate: (kind: "stores" | "products", id: string, status: "approved" | "rejected") => void;
}) {
  if (!items.length) return <p className="empty-copy">Nothing to review.</p>;
  return (
    <div className="record-list">
      {items.map((item) => {
        const status = "moderationStatus" in item ? item.moderationStatus : item.status;
        return (
          <div className="record-row moderation-row" key={item.id}>
            <div><strong>{item.name}</strong><span data-status={status}>{status}</span></div>
            <div className="row-actions">
              <button disabled={busy || status === "approved"} onClick={() => moderate(kind, item.id, "approved")}>Approve</button>
              <button disabled={busy || status === "rejected"} onClick={() => moderate(kind, item.id, "rejected")}>Reject</button>
            </div>
          </div>
        );
      })}
    </div>
  );
}
