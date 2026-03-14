export interface APIMeta {
  code: number;
  status: "success" | "error";
  message: string;
  request_id?: string;
  pagination?: APIPagination;
}

export interface APIPagination {
  page: number;
  limit: number;
  total_pages: number;
  total_records: number;
}

export interface APIErrorItem {
  field?: string;
  code: string;
  message: string;
}

export interface APIResponse<T> {
  meta: APIMeta;
  data: T;
}

export interface APIErrorResponse {
  meta: APIMeta;
  errors?: APIErrorItem[];
}
