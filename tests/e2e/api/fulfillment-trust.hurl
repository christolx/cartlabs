POST {{base_url}}/auth/login
Content-Type: application/json
{
  "email": "buyer3@demo.cartlabs.local",
  "password": "demo-pass-123"
}
HTTP 200
[Captures]
ft_buyer_token: jsonpath "$.accessToken"

PUT {{base_url}}/cart/items/01989f00-0000-7000-8000-000000000401
Authorization: Bearer {{ft_buyer_token}}
Content-Type: application/json
{
  "quantity": 1
}
HTTP 200

PUT {{base_url}}/cart/items/01989f00-0000-7000-8000-000000000402
Authorization: Bearer {{ft_buyer_token}}
Content-Type: application/json
{
  "quantity": 1
}
HTTP 200

POST {{base_url}}/checkout
Authorization: Bearer {{ft_buyer_token}}
Idempotency-Key: milestone-three-fulfillment-001
HTTP 201
[Captures]
ft_purchase_id: jsonpath "$.id"
ft_purchase_reference: jsonpath "$.reference"

POST {{base_url}}/purchases/{{ft_purchase_id}}/pay
Authorization: Bearer {{ft_buyer_token}}
Content-Type: application/json
{
  "outcome": "succeeded"
}
HTTP 200
[Asserts]
jsonpath "$.status" == "paid"

POST {{base_url}}/auth/login
Content-Type: application/json
{
  "email": "merchant@demo.cartlabs.local",
  "password": "demo-pass-123"
}
HTTP 200
[Captures]
ft_seller_one_token: jsonpath "$.accessToken"

GET {{base_url}}/seller/orders
Authorization: Bearer {{ft_seller_one_token}}
HTTP 200
[Captures]
ft_order_one_id: jsonpath "$.items[0].id"
ft_item_one_id: jsonpath "$.items[0].items[0].id"
[Asserts]
jsonpath "$.items[0].reference" == "{{ft_purchase_reference}}"
jsonpath "$.items[0].status" == "paid"

PATCH {{base_url}}/seller/orders/{{ft_order_one_id}}/status
Authorization: Bearer {{ft_seller_one_token}}
Content-Type: application/json
{
  "status": "processing",
  "reason": ""
}
HTTP 200
[Asserts]
jsonpath "$.status" == "processing"
jsonpath "$.processingAt" exists

PATCH {{base_url}}/seller/orders/{{ft_order_one_id}}/status
Authorization: Bearer {{ft_seller_one_token}}
Content-Type: application/json
{
  "status": "shipped",
  "reason": ""
}
HTTP 200
[Asserts]
jsonpath "$.status" == "shipped"
jsonpath "$.shippedAt" exists

PATCH {{base_url}}/seller/orders/{{ft_order_one_id}}/status
Authorization: Bearer {{ft_seller_one_token}}
Content-Type: application/json
{
  "status": "delivered",
  "reason": ""
}
HTTP 200
[Captures]
ft_delivered_item_id: jsonpath "$.items[0].id"
[Asserts]
jsonpath "$.status" == "delivered"
jsonpath "$.deliveredAt" exists

POST {{base_url}}/auth/login
Content-Type: application/json
{
  "email": "merchant2@demo.cartlabs.local",
  "password": "demo-pass-123"
}
HTTP 200
[Captures]
ft_seller_two_token: jsonpath "$.accessToken"

GET {{base_url}}/seller/orders
Authorization: Bearer {{ft_seller_two_token}}
HTTP 200
[Captures]
ft_order_two_id: jsonpath "$.items[0].id"
ft_item_two_id: jsonpath "$.items[0].items[0].id"

PATCH {{base_url}}/seller/orders/{{ft_order_one_id}}/status
Authorization: Bearer {{ft_seller_two_token}}
Content-Type: application/json
{
  "status": "cancelled",
  "reason": "Wrong owner"
}
HTTP 404

PATCH {{base_url}}/seller/orders/{{ft_order_two_id}}/status
Authorization: Bearer {{ft_seller_two_token}}
Content-Type: application/json
{
  "status": "cancelled",
  "reason": "Damaged before packing"
}
HTTP 200
[Asserts]
jsonpath "$.status" == "cancelled"
jsonpath "$.cancelledAt" exists
jsonpath "$.cancellationReason" == "Damaged before packing"

POST {{base_url}}/reviews
Authorization: Bearer {{ft_buyer_token}}
Content-Type: application/json
{
  "purchaseItemId": "{{ft_item_two_id}}",
  "rating": 1,
  "title": "Should be blocked",
  "body": "Cancelled order was never delivered."
}
HTTP 403

POST {{base_url}}/reviews
Authorization: Bearer {{ft_buyer_token}}
Content-Type: application/json
{
  "purchaseItemId": "{{ft_delivered_item_id}}",
  "rating": 5,
  "title": "Built for daily use",
  "body": "Verified delivery and sturdy finish."
}
HTTP 201
[Captures]
ft_review_id: jsonpath "$.id"
[Asserts]
jsonpath "$.rating" == 5
jsonpath "$.buyerName" == "Synthetic Traffic Buyer"

POST {{base_url}}/reviews
Authorization: Bearer {{ft_buyer_token}}
Content-Type: application/json
{
  "purchaseItemId": "{{ft_delivered_item_id}}",
  "rating": 4,
  "title": "Duplicate",
  "body": "Second review must not be accepted."
}
HTTP 409

GET {{base_url}}/catalog/products/handwoven-market-basket/reviews
HTTP 200
[Asserts]
jsonpath "$.average" == 5.0
jsonpath "$.count" == 1
jsonpath "$.items[0].id" == "{{ft_review_id}}"

PUT {{base_url}}/cart/items/01989f00-0000-7000-8000-000000000401
Authorization: Bearer {{ft_buyer_token}}
Content-Type: application/json
{
  "quantity": 1
}
HTTP 200

POST {{base_url}}/checkout
Authorization: Bearer {{ft_buyer_token}}
Idempotency-Key: milestone-three-cancel-001
HTTP 201
[Captures]
ft_cancel_purchase_id: jsonpath "$.id"

POST {{base_url}}/purchases/{{ft_cancel_purchase_id}}/cancel
Authorization: Bearer {{ft_buyer_token}}
Content-Type: application/json
{
  "reason": "Changed my mind"
}
HTTP 200
[Asserts]
jsonpath "$.status" == "cancelled"
jsonpath "$.paymentStatus" == "cancelled"
jsonpath "$.sellerOrders[0].status" == "cancelled"

GET {{base_url}}/catalog/products/handwoven-market-basket
HTTP 200
[Asserts]
jsonpath "$.variants[0].stock" == 17

POST {{base_url}}/auth/demo-login
Content-Type: application/json
{
  "role": "admin"
}
HTTP 200
[Captures]
ft_admin_token: jsonpath "$.accessToken"

GET {{base_url}}/admin/overview
Authorization: Bearer {{ft_admin_token}}
HTTP 200
[Asserts]
jsonpath "$.users" exists
jsonpath "$.purchases" exists
jsonpath "$.deliveredOrders" exists
jsonpath "$.grossMerchandiseMinor" exists
jsonpath "$.currency" == "IDR"

GET {{base_url}}/admin/audit-events
Authorization: Bearer {{ft_admin_token}}
HTTP 200
[Asserts]
jsonpath "$.items[0].action" == "purchase.cancelled"
jsonpath "$.items[0].actorRole" == "buyer"
jsonpath "$.items[1].action" == "review.created"

GET {{base_url}}/admin/audit-events
Authorization: Bearer {{ft_buyer_token}}
HTTP 403
