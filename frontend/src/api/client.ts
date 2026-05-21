import { ApiError, isApiErrorEnvelope } from "./error";

const BASE_URL = process.env.EXPO_PUBLIC_API_URL || "http://localhost:8080";

let getAuthToken: () => string | null = () => null;
let onUnauthorized: () => void = () => {};

export function configureApiClient(opts: {
  getAuthToken?: () => string | null;
  onUnauthorized?: () => void;
}) {
  if (opts.getAuthToken) getAuthToken = opts.getAuthToken;
  if (opts.onUnauthorized) onUnauthorized = opts.onUnauthorized;
}

export type FetchOptions = {
  url: string;
  method?: string;
  params?: Record<string, unknown>;
  headers?: Record<string, string>;
  data?: unknown;
  signal?: AbortSignal;
};

function buildUrl(path: string, params?: Record<string, unknown>): string {
  const url = path.startsWith("http") ? path : `${BASE_URL}${path}`;
  if (!params) return url;
  const search = new URLSearchParams();
  for (const [key, value] of Object.entries(params)) {
    if (value === undefined || value === null) continue;
    search.append(key, String(value));
  }
  const query = search.toString();
  return query ? `${url}?${query}` : url;
}

export async function vespinFetch<T>(url: string, options?: RequestInit): Promise<T>;
export async function vespinFetch<T>(options: FetchOptions): Promise<T>;
export async function vespinFetch<T>(input: string | FetchOptions, init?: RequestInit): Promise<T> {
  const generatedCall = typeof input === "string";
  const fetchOptions = generatedCall
    ? requestInitToFetchOptions(input, init)
    : input;
  const { url, method = "GET", params, headers, data, signal } = fetchOptions;

  const token = getAuthToken();
  const finalHeaders: Record<string, string> = {
    Accept: "application/json",
    ...(data !== undefined ? { "Content-Type": "application/json" } : {}),
    ...(token ? { Authorization: `Bearer ${token}` } : {}),
    ...headers,
  };

  const response = await fetch(buildUrl(url, params), {
    method,
    headers: finalHeaders,
    body: data !== undefined ? JSON.stringify(data) : undefined,
    signal,
  });

  if (response.status === 401) {
    onUnauthorized();
  }

  const text = await response.text();
  const body: unknown = text ? safeJsonParse(text) : undefined;

  if (!response.ok) {
    if (isApiErrorEnvelope(body)) {
      throw new ApiError(response.status, body.error.code, body.error.message);
    }
    throw new ApiError(response.status, "unknown_error", `Request failed with status ${response.status}`);
  }

  if (generatedCall) {
    return {
      data: body,
      headers: response.headers,
      status: response.status,
    } as T;
  }

  return body as T;
}

function safeJsonParse(text: string): unknown {
  try {
    return JSON.parse(text);
  } catch {
    return undefined;
  }
}

function requestInitToFetchOptions(url: string, init?: RequestInit): FetchOptions {
  const headers = headersToRecord(init?.headers);
  const data = typeof init?.body === "string" ? safeJsonParse(init.body) : init?.body;

  return {
    url,
    method: init?.method,
    headers,
    data,
    signal: init?.signal ?? undefined,
  };
}

function headersToRecord(headers?: HeadersInit): Record<string, string> | undefined {
  if (!headers) return undefined;
  if (headers instanceof Headers) return Object.fromEntries(headers.entries());
  if (Array.isArray(headers)) return Object.fromEntries(headers);
  return headers;
}

export default vespinFetch;
