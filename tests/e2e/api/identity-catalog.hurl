POST {{base_url}}/auth/demo-login
Content-Type: application/json
{
  "role": "seller"
}
HTTP 200
[Captures]
seller_token: jsonpath "$.accessToken"
[Asserts]
jsonpath "$.user.role" == "seller"
cookie "cartlabs_refresh" exists

GET {{base_url}}/seller/store
Authorization: Bearer {{seller_token}}
HTTP 200
[Captures]
store_id: jsonpath "$.id"
[Asserts]
jsonpath "$.slug" == "demo-seller-store"
jsonpath "$.status" == "approved"

POST {{base_url}}/auth/demo-login
Content-Type: application/json
{
  "role": "admin"
}
HTTP 200
[Captures]
admin_token: jsonpath "$.accessToken"

POST {{base_url}}/seller/products
Authorization: Bearer {{seller_token}}
Content-Type: application/json
{
  "categoryId": "01989f00-0000-7000-8000-000000000102",
  "name": "Milestone One Lamp",
  "slug": "milestone-one-lamp-{{newUuid}}",
  "description": "A warm desk lamp used to verify seller-to-buyer discovery."
}
HTTP 201
[Captures]
product_id: jsonpath "$.id"
product_slug: jsonpath "$.slug"
[Asserts]
jsonpath "$.status" == "draft"

POST {{base_url}}/seller/products/{{product_id}}/variants
Authorization: Bearer {{seller_token}}
Content-Type: application/json
{
  "sku": "M1-LAMP-BRASS",
  "name": "Brass",
  "attributes": {"finish": "Brass"},
  "priceMinor": 399000,
  "currency": "IDR",
  "stock": 7
}
HTTP 201
[Captures]
variant_id: jsonpath "$.id"
[Asserts]
jsonpath "$.stock" == 7

POST {{base_url}}/seller/products/{{product_id}}/images
Authorization: Bearer {{seller_token}}
Content-Type: application/json
{
  "url": "/images/shared-product.webp",
  "altText": "Brass desk lamp",
  "position": 0
}
HTTP 201

GET {{base_url}}/seller/products/{{product_id}}
Authorization: Bearer {{seller_token}}
HTTP 200
[Asserts]
jsonpath "$.id" == "{{product_id}}"
jsonpath "$.variants[0].id" == "{{variant_id}}"
jsonpath "$.images[0].position" == 0

POST {{base_url}}/auth/login
Content-Type: application/json
{
  "email": "merchant@demo.cartlabs.local",
  "password": "demo-pass-123"
}
HTTP 200
[Captures]
other_seller_token: jsonpath "$.accessToken"

GET {{base_url}}/seller/products/{{product_id}}
Authorization: Bearer {{other_seller_token}}
HTTP 404

POST {{base_url}}/seller/products/{{product_id}}/publish
Authorization: Bearer {{seller_token}}
HTTP 200
[Asserts]
jsonpath "$.status" == "published"

GET {{base_url}}/catalog/products?store=demo-seller-store&category=home-living&inStock=true
HTTP 200
[Asserts]
jsonpath "$.total" == 1
jsonpath "$.items[0].slug" == "{{product_slug}}"

PATCH {{base_url}}/admin/products/{{product_id}}/status
Authorization: Bearer {{admin_token}}
Content-Type: application/json
{
  "status": "suspended",
  "reason": "Listing enforcement workflow"
}
HTTP 200
[Asserts]
jsonpath "$.status" == "suspended"
jsonpath "$.enforcementReason" == "Listing enforcement workflow"

GET {{base_url}}/catalog/products?store=demo-seller-store&category=home-living&inStock=true
HTTP 200
[Asserts]
jsonpath "$.total" == 0

POST {{base_url}}/seller/products/{{product_id}}/publish
Authorization: Bearer {{seller_token}}
HTTP 409

PATCH {{base_url}}/admin/products/{{product_id}}/status
Authorization: Bearer {{admin_token}}
Content-Type: application/json
{
  "status": "published",
  "reason": "Seller corrected listing"
}
HTTP 200
[Asserts]
jsonpath "$.status" == "published"
jsonpath "$.enforcementReason" == "Listing enforcement workflow"

GET {{base_url}}/catalog/stores/demo-seller-store
HTTP 200
[Asserts]
jsonpath "$.name" == "Demo Seller Store"
jsonpath "$.sellerDisplayName" == "Demo Seller"

GET {{base_url}}/catalog/products?store=demo-seller-store&q=Milestone%20One%20Lamp&category=home-living&inStock=true
HTTP 200
[Asserts]
jsonpath "$.total" == 1
jsonpath "$.items[0].slug" == "{{product_slug}}"
jsonpath "$.items[0].minPriceMinor" == 399000
jsonpath "$.items[0].storeSlug" == "demo-seller-store"

GET {{base_url}}/catalog/products?store=nusantara-goods&q=Milestone%20One%20Lamp
HTTP 200
[Asserts]
jsonpath "$.total" == 0

GET {{base_url}}/catalog/products/{{product_slug}}
HTTP 200
[Asserts]
jsonpath "$.variants[0].stock" == 7
jsonpath "$.images[0].position" == 0
jsonpath "$.storeSlug" == "demo-seller-store"

PATCH {{base_url}}/seller/variants/{{variant_id}}/inventory
Authorization: Bearer {{seller_token}}
Content-Type: application/json
{
  "delta": -2,
  "reason": "Milestone workflow sale simulation"
}
HTTP 200
[Asserts]
jsonpath "$.stock" == 5

POST {{base_url}}/auth/demo-login
Content-Type: application/json
{
  "role": "buyer"
}
HTTP 200
[Captures]
buyer_token: jsonpath "$.accessToken"

GET {{base_url}}/admin/stores
Authorization: Bearer {{buyer_token}}
HTTP 403
