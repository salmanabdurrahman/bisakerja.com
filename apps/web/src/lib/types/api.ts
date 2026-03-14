/**
 * APIMeta defines the shape of api meta.
 */
export interface APIMeta {
  code: number;
  status: "success" | "error";
  message: string;
  request_id?: string;
  pagination?: APIPagination;
}

/**
 * APIPagination defines the shape of api pagination.
 */
export interface APIPagination {
  page: number;
  limit: number;
  total_pages: number;
  total_records: number;
}

/**
 * APIErrorItem defines the shape of api error item.
 */
export interface APIErrorItem {
  field?: string;
  code: string;
  message: string;
}

/**
 * APIResponse defines the shape of api response.
 */
export interface APIResponse<T> {
  meta: APIMeta;
  data: T;
}

/**
 * APIErrorResponse defines the shape of api error response.
 */
export interface APIErrorResponse {
  meta: APIMeta;
  errors?: APIErrorItem[];
}
