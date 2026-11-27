import { createContext } from 'react';

export type JsonRequest = <Result>(url: string, data?: unknown) => Promise<Result>;

const previewRequest: JsonRequest = async (url, data) => {
  const form = data instanceof FormData;
  const response = await fetch(url, {
    method: data === undefined ? 'GET' : 'POST',
    headers: data === undefined || form ? undefined : { 'Content-Type': 'application/json' },
    body: form ? data : JSON.stringify(data),
  });
  const payload = await response.json();
  if (!response.ok) throw Error(payload.error || '请求失败');
  return payload;
};

export const RequestContext = createContext<JsonRequest>(previewRequest);
