namespace go order

include "common.thrift"

enum OrderStatus {
    CREATED = 0
    PAID = 1
    SHIPPED = 2
    COMPLETED = 3
    CANCELLED = 4
}

struct CreateOrderRequest {
    1: required i64 user_id
    2: required i64 product_id
    3: required i32 quantity
    4: optional string address
}

struct CreateOrderResponse {
    1: string order_id
    2: OrderStatus status
    3: common.BaseResponse base_resp
}

struct GetOrderRequest {
    1: required string order_id
}

struct GetOrderResponse {
    1: string order_id
    2: i64 user_id
    3: i64 product_id
    4: i32 quantity
    5: string address
    6: OrderStatus status
    7: common.BaseResponse base_resp
}

struct ListOrdersRequest {
    1: required i64 user_id
    2: optional common.Pagination pagination
}

struct ListOrdersResponse {
    1: list<GetOrderResponse> orders
    2: common.Pagination pagination
    3: common.BaseResponse base_resp
}

exception OrderException {
    1: i32 code
    2: string message
}

service OrderService {
    CreateOrderResponse CreateOrder(1: CreateOrderRequest req) throws (1: OrderException e)
    GetOrderResponse GetOrder(1: GetOrderRequest req) throws (1: OrderException e)
    ListOrdersResponse ListOrders(1: ListOrdersRequest req)
}
