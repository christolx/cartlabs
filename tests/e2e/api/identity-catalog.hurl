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
HTTP 404

POST {{base_url}}/seller/store
Authorization: Bearer {{seller_token}}
Content-Type: application/json
{
  "name": "Demo Seller Studio",
  "slug": "demo-seller-studio",
  "description": "Catalog created by milestone end-to-end workflow."
}
HTTP 201
[Captures]
store_id: jsonpath "$.id"
[Asserts]
jsonpath "$.status" == "pending"

POST {{base_url}}/auth/demo-login
Content-Type: application/json
{
  "role": "admin"
}
HTTP 200
[Captures]
admin_token: jsonpath "$.accessToken"

PATCH {{base_url}}/admin/stores/{{store_id}}/moderation
Authorization: Bearer {{admin_token}}
Content-Type: application/json
{
  "status": "approved",
  "note": "Verified by milestone workflow"
}
HTTP 200
[Asserts]
jsonpath "$.status" == "approved"

POST {{base_url}}/seller/products
Authorization: Bearer {{seller_token}}
Content-Type: application/json
{
  "categoryId": "01989f00-0000-7000-8000-000000000102",
  "name": "Milestone One Lamp",
  "slug": "milestone-one-lamp",
  "description": "A warm desk lamp used to verify seller-to-buyer discovery."
}
HTTP 201
[Captures]
product_id: jsonpath "$.id"
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
jsonpath "$.moderationStatus" == "pending"

GET {{base_url}}/catalog/products?q=Milestone%20One%20Lamp
HTTP 200
[Asserts]
jsonpath "$.total" == 0

PATCH {{base_url}}/admin/products/{{product_id}}/moderation
Authorization: Bearer {{admin_token}}
Content-Type: application/json
{
  "status": "approved",
  "note": "Catalog quality verified"
}
HTTP 200
[Asserts]
jsonpath "$.moderationStatus" == "approved"

GET {{base_url}}/catalog/products?q=Milestone%20One%20Lamp&category=home-living&inStock=true
HTTP 200
[Asserts]
jsonpath "$.total" == 1
jsonpath "$.items[0].slug" == "milestone-one-lamp"
jsonpath "$.items[0].minPriceMinor" == 399000
jsonpath "$.items[0].storeSlug" == "demo-seller-studio"

GET {{base_url}}/catalog/stores/demo-seller-studio
HTTP 200
[Asserts]
jsonpath "$.name" == "Demo Seller Studio"
jsonpath "$.sellerDisplayName" == "Demo Seller"

GET {{base_url}}/catalog/products?store=demo-seller-studio&q=Milestone%20One%20Lamp&category=home-living&inStock=true
HTTP 200
[Asserts]
jsonpath "$.total" == 1
jsonpath "$.items[0].slug" == "milestone-one-lamp"

GET {{base_url}}/catalog/products?store=nusantara-goods&q=Milestone%20One%20Lamp
HTTP 200
[Asserts]
jsonpath "$.total" == 0

GET {{base_url}}/catalog/products/milestone-one-lamp
HTTP 200
[Asserts]
jsonpath "$.variants[0].stock" == 7
jsonpath "$.images[0].position" == 0
jsonpath "$.storeSlug" == "demo-seller-studio"

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
