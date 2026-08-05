GET {{base_url}}/health/live
HTTP 200
Content-Type: application/json
[Asserts]
jsonpath "$.status" == "ok"

GET {{base_url}}/health/ready
HTTP 200
Content-Type: application/json
[Asserts]
jsonpath "$.status" == "ok"
jsonpath "$.dependencies.postgres" == "ok"
jsonpath "$.dependencies.redis" == "ok"
jsonpath "$.dependencies.rabbitmq" == "ok"
