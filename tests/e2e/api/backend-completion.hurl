POST {{base_url}}/auth/demo-login
Content-Type: application/json
{
  "role": "admin"
}
HTTP 200
[Captures]
admin_token: jsonpath "$.accessToken"

GET {{base_url}}/admin/users
Authorization: Bearer {{admin_token}}
HTTP 200
[Asserts]
jsonpath "$.items" count > 0

POST {{base_url}}/auth/login
Content-Type: application/json
{
  "email": "buyer3@demo.cartlabs.local",
  "password": "demo-pass-123"
}
HTTP 200
[Captures]
old_buyer_token: jsonpath "$.accessToken"

PATCH {{base_url}}/admin/users/01989f00-0000-7000-8000-000000000007/status
Authorization: Bearer {{admin_token}}
Content-Type: application/json
{
  "status": "suspended",
  "reason": "Backend completion session enforcement check"
}
HTTP 200
[Asserts]
jsonpath "$.status" == "suspended"

GET {{base_url}}/me
Authorization: Bearer {{old_buyer_token}}
HTTP 401

POST {{base_url}}/auth/refresh
HTTP 401

PATCH {{base_url}}/admin/users/01989f00-0000-7000-8000-000000000007/status
Authorization: Bearer {{admin_token}}
Content-Type: application/json
{
  "status": "active",
  "reason": "Backend completion reactivation check"
}
HTTP 200
[Asserts]
jsonpath "$.status" == "active"

GET {{base_url}}/me
Authorization: Bearer {{old_buyer_token}}
HTTP 401

POST {{base_url}}/auth/login
Content-Type: application/json
{
  "email": "buyer3@demo.cartlabs.local",
  "password": "demo-pass-123"
}
HTTP 200
[Captures]
fresh_buyer_token: jsonpath "$.accessToken"

GET {{base_url}}/me
Authorization: Bearer {{fresh_buyer_token}}
HTTP 200

PATCH {{base_url}}/admin/users/01989f00-0000-7000-8000-000000000003/status
Authorization: Bearer {{admin_token}}
Content-Type: application/json
{
  "status": "suspended",
  "reason": "Self suspension must fail"
}
HTTP 409

GET {{base_url}}/admin/audit-events
Authorization: Bearer {{admin_token}}
HTTP 200
[Asserts]
jsonpath "$.items[0].action" == "user.status.active"
jsonpath "$.items[0].data.reason" == "Backend completion reactivation check"
