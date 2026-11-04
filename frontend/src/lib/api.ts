// REST client for the Go backend.
import type {
  BootstrapState,
  ConfigPatch,
  Profile,
  Group,
  RequestView,
  DnsQueryLine,
  LogLine,
  RunningState,
  ProviderView,
} from './types';

async function request<T>(method: string, url: string, body?: unknown): Promise<T> {
  const response = await fetch(url, {
    method,
    headers: body !== undefined ? { 'Content-Type': 'application/json' } : undefined,
    body: body !== undefined ? JSON.stringify(body) : undefined,
  });
  const text = await response.text();
  let data: any = null;
  try {
    data = text ? JSON.parse(text) : null;
  } catch {
    data = text;
  }
  if (!response.ok) {
    const message = data && typeof data === 'object' && data.error ? data.error : String(text || response.statusText);
    throw new Error(message);
  }
  return data as T;
}

export const api = {
  bootstrap: () => request<BootstrapState>('GET', '/api/bootstrap'),
  putConfig: (patch: ConfigPatch) => request<{ config: unknown; warning?: string }>('PUT', '/api/config', patch),
  applyConfig: () => request<{ ok: boolean }>('POST', '/api/config/apply'),

  profiles: () => request<Profile[]>('GET', '/api/profiles'),
  addProfile: (input: { url?: string; label?: string; type?: string; fileContent?: unknown; autoUpdate?: boolean; autoUpdateDuration?: number }) =>
    request<Profile>('POST', '/api/profiles', input),
  editProfile: (id: number, patch: Partial<Profile>) => request<Profile>('PUT', `/api/profiles/${id}`, patch),
  deleteProfile: (id: number) => request<{ ok: boolean }>('DELETE', `/api/profiles/${id}`),
  updateProfile: (id: number) => request<Profile>('POST', `/api/profiles/${id}/update`),
  selectProfile: (id: number) => request<{ ok: boolean }>('POST', `/api/profiles/${id}/select`),
  profileFile: (id: number) => request<{ content: string }>('GET', `/api/profiles/${id}/file`),
  saveProfileFile: (id: number, content: string) => request<{ ok: boolean }>('PUT', `/api/profiles/${id}/file`, { content }),

  proxies: () => request<{ groups: Group[]; delayMap: Record<string, Record<string, number>> }>('GET', '/api/proxies'),
  selectProxy: (groupName: string, proxyName: string) => request<{ ok: boolean }>('POST', '/api/proxies/select', { 'group-name': groupName, 'proxy-name': proxyName }),
  testDelay: (target: { group?: string; proxy?: string }) => request<{ ok: boolean }>('POST', '/api/proxies/delay', target),

  requests: (feed: 'live' | 'recent') => request<RequestView[]>('GET', `/api/requests?feed=${feed}`),
  closeRequest: (id: string) => request<{ ok: boolean }>('DELETE', `/api/requests/${id}`),
  closeAllRequests: () => request<{ ok: boolean }>('DELETE', '/api/requests'),
  dnsQueries: () => request<DnsQueryLine[]>('GET', '/api/dns'),
  clearDnsCache: () => request<{ ok: boolean }>('POST', '/api/dns/clear'),
  logs: () => request<LogLine[]>('GET', '/api/logs'),
  resetTraffic: () => request<{ ok: boolean }>('POST', '/api/traffic/reset'),

  coreStart: () => request<RunningState>('POST', '/api/core/start'),
  coreStop: () => request<RunningState>('POST', '/api/core/stop'),
  coreRestart: () => request<RunningState>('POST', '/api/core/restart'),
  coreMemory: () => request<{ rss: number; heapInuse: number; heapIdle: number; stackInuse: number; runtimeOther: number }>('GET', '/api/core/memory'),
  probeIP: () => request<unknown>('POST', '/api/probe/ip', {}),
  ipQuality: (ip: string) => request<{ type: string; organization?: string; asn?: string; level: string }>('GET', `/api/ip-quality/${encodeURIComponent(ip)}`),
  probeServices: (names: string[]) => request<unknown>('POST', '/api/probe/services', { names }),
  geoUpdate: (type: string) => request<{ ok: boolean }>('POST', '/api/geo/update', { type }),

  providers: () => request<Record<string, ProviderView>>('GET', '/api/providers'),
  updateProvider: (name: string) => request<{ ok: boolean }>('POST', `/api/providers/${encodeURIComponent(name)}/update`),
  sideloadProvider: (name: string, data: string) => request<{ ok: boolean }>('POST', `/api/providers/${encodeURIComponent(name)}/sideload`, { data }),

  shareDecode: (lines: string[]) => request<{ proxies: { name: string; type: string }[] }>('POST', '/api/tools/share-decode', { lines }),
};
