import { describe, expect, it } from 'vitest';
import { ApiError, unwrapResponse } from './http';

describe('unwrapResponse', () => {
  it('解析 user-svc 成功包装格式', () => {
    expect(unwrapResponse({ success: true, data: { id: 1 } })).toEqual({ id: 1 });
  });

  it('success=false 时抛出 ApiError', () => {
    expect(() => unwrapResponse({ success: false, error: '用户名已存在' })).toThrow('用户名已存在');
  });

  it('无 success 字段时原样返回（plugin-svc 格式）', () => {
    expect(unwrapResponse({ plugins: [] })).toEqual({ plugins: [] });
  });
});

describe('ApiError', () => {
  it('保留 HTTP 状态码', () => {
    const err = new ApiError(401, '未授权');
    expect(err.status).toBe(401);
    expect(err.name).toBe('ApiError');
  });
});
