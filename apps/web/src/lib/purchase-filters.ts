import type { components } from "@/lib/api/schema";

export type Purchase = components["schemas"]["Purchase"];
export type PurchaseFilter =
  "all" | "awaiting" | "active" | "delivered" | "cancelled";

export const purchaseFilters: { value: PurchaseFilter; label: string }[] = [
  { value: "all", label: "All" },
  { value: "awaiting", label: "Awaiting payment" },
  { value: "active", label: "Active" },
  { value: "delivered", label: "Delivered" },
  { value: "cancelled", label: "Cancelled" },
];

export function purchaseItems(purchase: Purchase) {
  return purchase.sellerOrders.flatMap((order) => order.items);
}

export function itemCount(purchase: Purchase) {
  return purchaseItems(purchase).reduce(
    (total, item) => total + item.quantity,
    0,
  );
}

export function fulfillmentLabel(purchase: Purchase) {
  const delivered = purchase.sellerOrders.filter(
    (order) => order.status === "delivered",
  ).length;
  return `${delivered}/${purchase.sellerOrders.length} delivered`;
}

export function matchesPurchaseFilter(
  purchase: Purchase,
  filter: PurchaseFilter,
) {
  if (filter === "all") return true;
  if (filter === "awaiting") return purchase.status === "pending_payment";

  const sellerOrders = purchase.sellerOrders;
  const fullyDelivered =
    sellerOrders.length > 0 &&
    sellerOrders.every((order) => order.status === "delivered");
  const fullyCancelled =
    sellerOrders.length > 0 &&
    sellerOrders.every((order) => order.status === "cancelled");

  if (filter === "delivered") return fullyDelivered;
  if (filter === "cancelled") {
    return (
      ["cancelled", "expired", "payment_failed"].includes(purchase.status) ||
      fullyCancelled
    );
  }

  const hasLiveSellerOrder = sellerOrders.some((order) =>
    ["paid", "processing", "shipped"].includes(order.status),
  );
  return purchase.status === "paid" && hasLiveSellerOrder && !fullyDelivered;
}

export function matchesPurchaseQuery(purchase: Purchase, query: string) {
  const normalized = query.trim().toLocaleLowerCase();
  if (!normalized) return true;
  const searchable = [
    purchase.reference,
    ...purchase.sellerOrders.flatMap((order) => [
      order.reference,
      order.storeName,
      ...order.items.flatMap((item) => [
        item.productName,
        item.variantName,
        item.sku,
      ]),
    ]),
  ]
    .join(" ")
    .toLocaleLowerCase();
  return searchable.includes(normalized);
}
