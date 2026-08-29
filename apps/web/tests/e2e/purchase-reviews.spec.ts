import { expect, test, type Page, type Route } from "@playwright/test";

const buyer = {
  id: "11111111-1111-4111-8111-111111111111",
  email: "buyer@demo.cartlabs.local",
  displayName: "Demo Buyer",
  role: "buyer",
};

const purchaseId = "44444444-4444-4444-8444-444444444444";
const itemId = "55555555-5555-4555-8555-555555555555";

function purchase(reviewed: boolean) {
  return {
    id: purchaseId,
    reference: "CL-REVIEW000001",
    buyerId: buyer.id,
    status: "paid",
    paymentStatus: "succeeded",
    paymentIntentId: "pi_review",
    currency: "IDR",
    subtotalMinor: 249000,
    totalMinor: 249000,
    reservationExpiresAt: "2026-08-01T00:00:00Z",
    sellerOrders: [
      {
        id: "66666666-6666-4666-8666-666666666666",
        purchaseId,
        reference: "CL-REVIEW000001",
        storeId: "77777777-7777-4777-8777-777777777777",
        storeName: "Nusantara Goods",
        status: "delivered",
        currency: "IDR",
        subtotalMinor: 249000,
        items: [
          {
            id: itemId,
            productId: "88888888-8888-4888-8888-888888888888",
            variantId: "99999999-9999-4999-8999-999999999999",
            productName: "Handwoven Market Basket",
            variantName: "Natural",
            sku: "NUSA-BASKET-NAT",
            imageUrl: "/images/catalog-v2/loom-carry-basket.webp",
            quantity: 1,
            unitPriceMinor: 249000,
            lineTotalMinor: 249000,
            currency: "IDR",
            reviewed,
          },
        ],
        deliveredAt: "2026-08-05T00:00:00Z",
        cancellationReason: "",
        createdAt: "2026-08-01T00:00:00Z",
        updatedAt: "2026-08-05T00:00:00Z",
      },
    ],
    createdAt: "2026-08-01T00:00:00Z",
    updatedAt: "2026-08-05T00:00:00Z",
  };
}

async function json(route: Route, body: unknown, status = 200) {
  await route.fulfill({
    status,
    contentType: "application/json",
    body: JSON.stringify(body),
  });
}

async function mockBuyer(
  page: Page,
  reviewHandler: (route: Route) => Promise<void>,
) {
  await page.route("**/api/backend/**", async (route) => {
    const url = new URL(route.request().url());
    if (url.pathname.endsWith("/auth/refresh")) {
      return json(route, {
        accessToken: "token-buyer",
        tokenType: "Bearer",
        expiresIn: 900,
        user: buyer,
      });
    }
    if (url.pathname.endsWith("/me")) return json(route, buyer);
    if (url.pathname.endsWith("/reviews")) return reviewHandler(route);
    return json(route, { title: "Not found", status: 404 }, 404);
  });
}

test("reviewed purchase item stays hidden after reload", async ({ page }) => {
  await mockBuyer(page, async (route) =>
    json(route, { title: "Conflict", status: 409 }, 409),
  );
  await page.route(`**/api/backend/purchases/${purchaseId}`, async (route) =>
    json(route, purchase(true)),
  );

  await page.goto(`/purchases/${purchaseId}`);
  await expect(
    page.getByRole("heading", { name: "Handwoven Market Basket" }),
  ).toBeVisible();
  await expect(page.getByText("Write review")).toHaveCount(0);
  await page.reload();
  await expect(page.getByText("Write review")).toHaveCount(0);
});

test("eligible purchase item can be reviewed once", async ({ page }) => {
  let reviewed = false;
  await mockBuyer(page, async (route) => {
    reviewed = true;
    return json(
      route,
      {
        id: "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa",
        buyerId: buyer.id,
        buyerName: buyer.displayName,
        purchaseItemId: itemId,
        productId: "88888888-8888-4888-8888-888888888888",
        rating: 5,
        title: "Built for market days",
        body: "Strong weave and comfortable handles.",
        createdAt: "2026-08-06T00:00:00Z",
        updatedAt: "2026-08-06T00:00:00Z",
      },
      201,
    );
  });
  await page.route(`**/api/backend/purchases/${purchaseId}`, async (route) =>
    json(route, purchase(reviewed)),
  );

  await page.goto(`/purchases/${purchaseId}`);
  await page.getByText("Write review").click();
  await page.getByLabel("Title").fill("Built for market days");
  await page.getByLabel("Review").fill("Strong weave and comfortable handles.");
  await page.getByRole("button", { name: "Publish review" }).click();
  await expect(page.getByText("Verified review published.")).toBeVisible();
  await expect(page.getByText("Write review")).toHaveCount(0);
  await page.reload();
  await expect(page.getByText("Write review")).toHaveCount(0);
});

test("duplicate review conflict gets clear message and refreshed state", async ({
  page,
}) => {
  let reviewed = false;
  await mockBuyer(page, async (route) => {
    reviewed = true;
    return json(
      route,
      {
        title: "Conflict",
        detail: "resource state conflicts with request",
        status: 409,
      },
      409,
    );
  });
  await page.route(`**/api/backend/purchases/${purchaseId}`, async (route) =>
    json(route, purchase(reviewed)),
  );

  await page.goto(`/purchases/${purchaseId}`);
  await page.getByText("Write review").click();
  await page.getByLabel("Title").fill("Already sent");
  await page
    .getByLabel("Review")
    .fill("Duplicate submission after stale page.");
  await page.getByRole("button", { name: "Publish review" }).click();
  await expect(
    page.getByText("This item has already been reviewed."),
  ).toBeVisible();
  await expect(page.getByText("Write review")).toHaveCount(0);
});
