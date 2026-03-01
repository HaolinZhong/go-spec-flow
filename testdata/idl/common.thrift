namespace go common

struct BaseResponse {
    1: i32 code
    2: string message
}

struct Pagination {
    1: i32 page
    2: i32 page_size
    3: i64 total
}

enum StatusCode {
    SUCCESS = 0
    INVALID_PARAM = 1
    NOT_FOUND = 2
    INTERNAL_ERROR = 3
}
