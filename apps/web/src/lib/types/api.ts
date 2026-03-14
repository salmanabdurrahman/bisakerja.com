export interface APIMeta {
  code: number;
  status: "success" | "error";
  message: string;
  request_id?: string;
}

export interface APIResponse<T> {
  meta: APIMeta;
  data: T;
}
