import { describe, expect, it } from "vitest";
import type { components } from "@/lib/api/schema";
import {
  matchesPurchaseFilter,
  matchesPurchaseQuery,
  type Purchase,
} from "./purchase-filters";

type SellerOrderStatus = components["schemas"]["SellerOrderStatus"];

function purchase(
  statuses: SellerOrderStatus[],
  status: Purchase["status"] = "paid",
): Purchase {
  return {
    id: "purchase-1",
    reference: "CL-01989F000000",
    buyerId: "buyer-1",
    status,
    paymentStatus: status === "paid" ? "succeeded" : "pending",
    paymentIntentId: "intent-1",
    currency: "IDR",
    subtotalMinor: 100,
    totalMinor: 100,
    reservationExpiresAt: "2026-08-31T00:00:00Z",
    sellerOrders: statuses.map((orderStatus, index) => ({
      id: `order-${index}`,
      purchaseId: "purchase-1",
      reference: `CL-ORDER-${index}`,
      storeId: `store-${index}`,
      storeName: index === 0 ? "Demo Seller Store" : "Second Store",
      status: orderStatus,
      currency: "IDR",
      subtotalMinor: 100,
      items: [
        {
          id: `item-${index}`,
          productId: "product-1",
          variantId: "variant-1",
          productName: "Roll-Top Backpack",
          variantName: "Navy",
          sku: "TEN-BACKPACK-NAVY",
          imageUrl: "",
          quantity: 1,
          unitPriceMinor: 100,
          lineTotalMinor: 100,
          currency: "IDR",
          reviewed: false,
        },
      ],
      cancellationReason: "",
      createdAt: "2026-08-31T00:00:00Z",
      updatedAt: "2026-08-31T00:00:00Z",
    })),
    createdAt: "2026-08-31T00:00:00Z",
    updatedAt: "2026-08-31T00:00:00Z",
  };
}

describe("purchase filters", () => {
  it("keeps delivered purchases out of Active", () => {
    const delivered = purchase(["delivered"]);

    expect(matchesPurchaseFilter(delivered, "delivered")).toBe(true);
    expect(matchesPurchaseFilter(delivered, "active")).toBe(false);
  });

  it("keeps mixed fulfillment purchases in Active", () => {
    const mixed = purchase(["delivered", "shipped"]);

    expect(matchesPurchaseFilter(mixed, "active")).toBe(true);
    expect(matchesPurchaseFilter(mixed, "delivered")).toBe(false);
  });

  it("matches product, store, and reference search text", () => {
    const item = purchase(["paid"]);

    expect(matchesPurchaseQuery(item, "backpack")).toBe(true);
    expect(matchesPurchaseQuery(item, "seller store")).toBe(true);
    expect(matchesPurchaseQuery(item, "does-not-exist")).toBe(false);
  });
});
