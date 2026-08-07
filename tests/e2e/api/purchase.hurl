POST {{base_url}}/auth/demo-login
Content-Type: application/json
{
  "role": "buyer"
}
HTTP 200
[Captures]
buyer_token: jsonpath "$.accessToken"

PUT {{base_url}}/cart/items/01989f00-0000-7000-8000-000000000401
Authorization: Bearer {{buyer_token}}
Content-Type: application/json
{
  "quantity": 1
}
HTTP 200
[Asserts]
jsonpath "$.totalQuantity" == 1
jsonpath "$.subtotalMinor" == 249000

PUT {{base_url}}/cart/items/01989f00-0000-7000-8000-000000000402
Authorization: Bearer {{buyer_token}}
Content-Type: application/json
{
  "quantity": 1
}
HTTP 200
[Asserts]
jsonpath "$.totalQuantity" == 2
jsonpath "$.subtotalMinor" == 408000
jsonpath "$.stores[0].storeId" exists
jsonpath "$.stores[1].storeId" exists

POST {{base_url}}/checkout
Authorization: Bearer {{buyer_token}}
Idempotency-Key: milestone-two-multi-seller-001
HTTP 201
[Captures]
purchase_id: jsonpath "$.id"
purchase_reference: jsonpath "$.reference"
[Asserts]
jsonpath "$.status" == "pending_payment"
jsonpath "$.paymentStatus" == "pending"
jsonpath "$.paymentIntentId" exists
jsonpath "$.totalMinor" == 408000
jsonpath "$.sellerOrders[0].id" exists
jsonpath "$.sellerOrders[1].id" exists

POST {{base_url}}/checkout
Authorization: Bearer {{buyer_token}}
Idempotency-Key: milestone-two-multi-seller-001
HTTP 201
[Asserts]
jsonpath "$.id" == "{{purchase_id}}"
jsonpath "$.totalMinor" == 408000

GET {{base_url}}/cart
Authorization: Bearer {{buyer_token}}
HTTP 200
[Asserts]
jsonpath "$.totalQuantity" == 0
jsonpath "$.subtotalMinor" == 0

GET {{base_url}}/purchases/{{purchase_id}}
Authorization: Bearer {{buyer_token}}
HTTP 200
[Asserts]
jsonpath "$.reference" == "{{purchase_reference}}"
jsonpath "$.sellerOrders[0].items[0].unitPriceMinor" exists
jsonpath "$.sellerOrders[1].items[0].unitPriceMinor" exists

POST {{base_url}}/purchases/{{purchase_id}}/pay
Authorization: Bearer {{buyer_token}}
Content-Type: application/json
{
  "outcome": "succeeded"
}
HTTP 200
[Asserts]
jsonpath "$.status" == "paid"
jsonpath "$.paymentStatus" == "succeeded"
jsonpath "$.sellerOrders[0].status" == "paid"
jsonpath "$.sellerOrders[1].status" == "paid"

POST {{base_url}}/purchases/{{purchase_id}}/pay
Authorization: Bearer {{buyer_token}}
Content-Type: application/json
{
  "outcome": "succeeded"
}
HTTP 409

GET {{base_url}}/notifications
Authorization: Bearer {{buyer_token}}
HTTP 200
[Asserts]
jsonpath "$.items[0].kind" == "purchase.paid"
jsonpath "$.items[1].kind" == "purchase.created"

POST {{base_url}}/auth/login
Content-Type: application/json
{
  "email": "merchant@demo.cartlabs.local",
  "password": "demo-pass-123"
}
HTTP 200
[Captures]
seller_one_token: jsonpath "$.accessToken"

GET {{base_url}}/seller/orders
Authorization: Bearer {{seller_one_token}}
HTTP 200
[Asserts]
jsonpath "$.items[0].reference" == "{{purchase_reference}}"
jsonpath "$.items[0].status" == "paid"
jsonpath "$.items[0].subtotalMinor" == 249000

GET {{base_url}}/notifications
Authorization: Bearer {{seller_one_token}}
HTTP 200
[Asserts]
jsonpath "$.items[0].kind" == "purchase.paid"

POST {{base_url}}/auth/login
Content-Type: application/json
{
  "email": "merchant2@demo.cartlabs.local",
  "password": "demo-pass-123"
}
HTTP 200
[Captures]
seller_two_token: jsonpath "$.accessToken"

GET {{base_url}}/seller/orders
Authorization: Bearer {{seller_two_token}}
HTTP 200
[Asserts]
jsonpath "$.items[0].reference" == "{{purchase_reference}}"
jsonpath "$.items[0].status" == "paid"
jsonpath "$.items[0].subtotalMinor" == 159000

GET {{base_url}}/seller/orders
Authorization: Bearer {{buyer_token}}
HTTP 403

POST {{base_url}}/auth/login
Content-Type: application/json
{
  "email": "buyer2@demo.cartlabs.local",
  "password": "demo-pass-123"
}
HTTP 200
[Captures]
buyer_two_token: jsonpath "$.accessToken"

PUT {{base_url}}/cart/items/01989f00-0000-7000-8000-000000000401
Authorization: Bearer {{buyer_two_token}}
Content-Type: application/json
{
  "quantity": 1
}
HTTP 200
[Asserts]
jsonpath "$.subtotalMinor" == 249000

POST {{base_url}}/checkout
Authorization: Bearer {{buyer_two_token}}
Idempotency-Key: milestone-two-failed-payment-001
HTTP 201
[Captures]
failed_purchase_id: jsonpath "$.id"
[Asserts]
jsonpath "$.status" == "pending_payment"

POST {{base_url}}/purchases/{{failed_purchase_id}}/pay
Authorization: Bearer {{buyer_two_token}}
Content-Type: application/json
{
  "outcome": "failed"
}
HTTP 200
[Asserts]
jsonpath "$.status" == "payment_failed"
jsonpath "$.paymentStatus" == "failed"
jsonpath "$.sellerOrders[0].status" == "cancelled"

GET {{base_url}}/catalog/products/handwoven-market-basket
HTTP 200
[Asserts]
jsonpath "$.variants[0].stock" == 17

GET {{base_url}}/notifications
Authorization: Bearer {{buyer_two_token}}
HTTP 200
[Asserts]
jsonpath "$.items[0].kind" == "purchase.payment_failed"
jsonpath "$.items[1].kind" == "purchase.created"

POST {{base_url}}/payments/webhook
Content-Type: application/json
X-Cartlabs-Signature: t=0,v1=deadbeef
{
  "id": "01989f00-0000-7000-8000-000000000999",
  "type": "payment.succeeded",
  "createdAt": "2026-08-08T00:00:00Z",
  "data": {
    "intentId": "01989f00-0000-7000-8000-000000000998",
    "reference": "CL-01989F000000",
    "amountMinor": 1,
    "currency": "IDR"
  }
}
HTTP 401
