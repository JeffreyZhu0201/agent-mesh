/** API 网关基础地址，可通过 VITE_API_BASE_URL 覆盖 */
export const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080/api';

/** 带 HTTP 状态码的 API 错误，便于上层区分 401/409 等场景 */
export class ApiError extends Error {
  constructor(
    public status: number,
    message: string
  ) {
    super(message);
    this.name = 'ApiError';
  }
}

/** 兼容旧名称，auth 模块仍导出 AuthApiError */
export { ApiError as AuthApiError };

type JsonBody = Record<string, unknown>;

/** 从响应 JSON 中提取错误信息（兼容 user-svc 与 plugin-svc 两种格式） */
function extractErrorMessage(body: JsonBody, fallback: string): string {
  const msg = body.error ?? body.message;
  return typeof msg === 'string' ? msg : fallback;
}

/**
 * 解析后端响应体。
 * - user-svc：{ success, data } / { success: false, error }
 * - plugin-svc：直接返回业务对象（如 { plugins: [...] }）
 */
export function unwrapResponse<T>(body: unknown): T {
  if (body && typeof body === 'object' && 'success' in body) {
    const wrapped = body as { success: boolean; data?: T; error?: string };
    if (!wrapped.success) {
      throw new ApiError(400, wrapped.error || 'Request failed');
    }
    return wrapped.data as T;
  }
  return body as T;
}

/** 发起 HTTP 请求并解析 JSON，失败时抛出 ApiError */
export async function fetchJson<T>(
  url: string,
  options: RequestInit = {}
): Promise<{ ok: boolean; status: number; data: T }> {
  const headers = new Headers(options.headers);
  if (!headers.has('Content-Type') && options.body) {
    headers.set('Content-Type', 'application/json');
  }

  const response = await fetch(url, { ...options, headers });
  const body = (await response.json().catch(() => ({}))) as JsonBody;

  if (!response.ok) {
    throw new ApiError(response.status, extractErrorMessage(body, 'Request failed'));
  }

  return { ok: true, status: response.status, data: unwrapResponse<T>(body) };
}
