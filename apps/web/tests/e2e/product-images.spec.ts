import { expect, test, type Page, type Route } from "@playwright/test";

const seller = { id: "22222222-2222-4222-8222-222222222222", email: "seller@demo.cartlabs.local", displayName: "Demo Seller", role: "seller" } as const;
const imageFile = "public/images/catalog-v2/compact-digital-camera.webp";

type ProductImage = { id: string; url: string; altText: string; position: number };

function product(images: ProductImage[] = [], status: "draft" | "published" = "draft") {
  return {
    id: "11111111-1111-4111-8111-111111111111",
    storeId: "22222222-2222-4222-8222-222222222222",
    storeSlug: "demo-seller-store",
    category: { id: "33333333-3333-4333-8333-333333333333", name: "Home & Living", slug: "home-living" },
    name: "Upload Lamp",
    slug: "upload-lamp",
    description: "Product used to verify image lifecycle.",
    status,
    variants: [{ id: "44444444-4444-4444-8444-444444444444", sku: "LAMP-001", name: "Standard", attributes: {}, priceMinor: 399000, currency: "IDR", stock: 7, active: true }],
    images,
    createdAt: "2026-08-20T00:00:00Z",
    updatedAt: "2026-08-20T00:00:00Z",
  };
}

async function json(route: Route, body: unknown, status = 200) {
  await route.fulfill({ status, contentType: "application/json", body: JSON.stringify(body) });
}

async function mockProductPage(page: Page, initialImages: ProductImage[] = [], initialStatus: "draft" | "published" = "draft") {
  let images = initialImages;
  await page.route("**/api/cloudinary-config", (route) => json(route, { cloudName: "demo-cloud", uploadPreset: "cartlabs-upload" }));
  await page.route("https://api.cloudinary.com/**", (route) => json(route, { secure_url: "https://res.cloudinary.com/demo-cloud/image/upload/v1/lamp.webp" }));
  await page.route("**/api/backend/**", async (route) => {
    const url = new URL(route.request().url());
    if (url.pathname.endsWith("/auth/refresh")) return json(route, { accessToken: "seller-token", tokenType: "Bearer", expiresIn: 900, user: seller });
    if (url.pathname.endsWith("/me")) return json(route, seller);
    if (url.pathname.endsWith("/categories")) return json(route, { items: [product().category] });
    if (url.pathname.endsWith("/seller/products/11111111-1111-4111-8111-111111111111") && route.request().method() === "GET") return json(route, product(images, initialStatus));
    if (url.pathname.endsWith("/images") && route.request().method() === "POST") {
      images = [{ id: "55555555-5555-4555-8555-555555555555", url: "https://res.cloudinary.com/demo-cloud/image/upload/v1/lamp.webp", altText: "Desk lamp", position: 0 }];
      return json(route, images[0], 201);
    }
    if (url.pathname.includes("/images/") && route.request().method() === "PUT") {
      images = [{ ...images[0], url: "https://res.cloudinary.com/demo-cloud/image/upload/v2/lamp.webp", altText: "Replacement lamp" }];
      return json(route, images[0]);
    }
    if (url.pathname.includes("/images/") && route.request().method() === "DELETE") {
      images = [];
      return route.fulfill({ status: 204 });
    }
    return json(route, { title: "Not found", status: 404 }, 404);
  });
}

test("seller uploads, replaces, and deletes product image", async ({ page }) => {
  await mockProductPage(page);
  await page.goto("/seller/products/11111111-1111-4111-8111-111111111111");
  await expect(page.getByRole("heading", { name: "Product images" })).toBeVisible();

  await page.getByRole("button", { name: "Product image file" }).setInputFiles(imageFile);
  await expect(page.getByAltText("Selected image preview")).toBeVisible();
  await expect(page.getByRole("button", { name: "Add image" })).toBeEnabled();
  await page.getByLabel("Alt text").fill("Desk lamp");
  await page.getByRole("button", { name: "Add image" }).click();
  await expect(page.getByRole("status")).toHaveText("Image added.");

  await page.locator(".image-actions input[type=file]").setInputFiles(imageFile);
  await expect(page.getByRole("status")).toHaveText("Image replaced.");
  page.once("dialog", (dialog) => dialog.accept());
  await page.getByRole("button", { name: "Delete" }).click();
  await expect(page.getByRole("status")).toHaveText("Image deleted.");
  await expect(page.getByText("No images uploaded.")).toBeVisible();
});

test("seller cannot add ninth image or delete only published image", async ({ page }) => {
  const image = { id: "55555555-5555-4555-8555-555555555555", url: "/images/shared-product.webp", altText: "Desk lamp", position: 0 };
  await mockProductPage(page, [image], "published");
  await page.goto("/seller/products/11111111-1111-4111-8111-111111111111");
  await expect(page.getByRole("button", { name: "Delete" })).toBeDisabled();

  const eightImages = Array.from({ length: 8 }, (_, position) => ({ ...image, id: `55555555-5555-4555-8555-55555555555${position}`, position }));
  await mockProductPage(page, eightImages);
  await page.reload();
  await expect(page.getByText("8 / 8")).toBeVisible();
  await expect(page.getByRole("button", { name: "Eight-image limit reached" })).toBeDisabled();
  await expect(page.getByRole("button", { name: "Product image file" })).toBeDisabled();
});
