// Central client state, fed by the bootstrap request and the WebSocket.
import { writable, derived, get } from 'svelte/store';
import { api } from './api';
import type {
  BootstrapState,
  Config,
  ConfigPatch,
  Group,
  Profile,
  RequestView,
  DnsQueryLine,
  LogLine,
  RunningState,
  ProviderView,
  DelayEvent,
  TrafficEvent,
} from './types';

export const config = writable<Config | null>(null);
export const profiles = writable<Profile[]>([]);
export const groups = writable<Group[]>([]);
export const delayMap = writable<Record<string, Record<string, number>>>({});
export const running = writable<RunningState>({ running: false, startTime: 0, systemProxy: false, port: 7890, mode: 'rule' });
export const providers = writable<Record<string, ProviderView>>({});
export const liveRequests = writable<RequestView[]>([]);
export const recentRequests = writable<RequestView[]>([]);
export const liveCount = writable(0);
export const dnsQueries = writable<DnsQueryLine[]>([]);
export const logs = writable<LogLine[]>([]);
export const trafficHistory = writable<{ up: number; down: number }[]>([]);
export const totalTraffic = writable({ up: 0, down: 0 });
export const connected = writable(false);
export const coreConnected = writable(false);
export const banners = writable<{ id: number; level: string; message: string }[]>([]);

let bannerSeq = 0;

export function showBanner(message: string, level: 'info' | 'error' | 'success' = 'error') {
  const id = ++bannerSeq;
  banners.update((list) => [...list, { id, level, message }]);
  setTimeout(() => banners.update((list) => list.filter((b) => b.id !== id)), 6000);
}

// currentProfile derived from config + profiles
export const currentProfile = derived([config, profiles], ([$config, $profiles]) => {
  if (!$config?.currentProfileId) return null;
  return $profiles.find((p) => p.id === $config.currentProfileId) ?? null;
});

export const selectedMap = derived(currentProfile, ($p) => $p?.selectedMap ?? {});

export const defaultTestUrl = derived(config, ($c) => $c?.patchClashConfig['test-url'] ?? 'https://www.gstatic.com/generate_204');

// The tunnel mode lives in the patch config; the UI reads it from there so a
// mode click flips instantly from the PUT response.
export const tunnelMode = derived([config, running], ([$c, $r]) => $c?.patchClashConfig['mode'] ?? $r.mode);

// applyEvent folds one WS event into the stores.
export function applyEvent(type: string, data: any) {
  switch (type) {
    case 'init': {
      const snapshot = data as {
        logs: LogLine[];
        dns: DnsQueryLine[];
        live: RequestView[];
        recent: RequestView[];
        liveCount: number;
      };
      logs.set(snapshot.logs ?? []);
      dnsQueries.set(snapshot.dns ?? []);
      liveRequests.set(snapshot.live ?? []);
      recentRequests.set(snapshot.recent ?? []);
      liveCount.set(snapshot.liveCount ?? 0);
      break;
    }
    case 'config':
      config.set(data as Config);
      break;
    case 'groups':
      groups.set(data as Group[]);
      break;
    case 'delay': {
      const event = data as DelayEvent;
      delayMap.update((map) => {
        const url = event.url || get(defaultTestUrl);
        const next = { ...map, [url]: { ...(map[url] ?? {}), [event.name]: event.value } };
        return next;
      });
      break;
    }
    case 'request': {
      const view = data as RequestView;
      liveRequests.update((list) => {
        if (list.some((r) => r.id === view.id)) return list;
        return [...list, view];
      });
      recentRequests.update((list) => [view, ...list].slice(0, 3000));
      break;
    }
    case 'requestTraffic': {
      const updates = data as { id: string; upload: number; download: number }[];
      const byId = new Map(updates.map((u) => [u.id, u]));
      liveRequests.update((list) =>
        list.map((r) => {
          const u = byId.get(r.id);
          return u ? { ...r, upload: u.upload, download: u.download } : r;
        }),
      );
      recentRequests.update((list) =>
        list.map((r) => {
          const u = byId.get(r.id);
          return u ? { ...r, upload: u.upload, download: u.download } : r;
        }),
      );
      break;
    }
    case 'requestCount':
      liveCount.set(data as number);
      break;
    case 'dns':
      dnsQueries.update((list) => [...list, data as DnsQueryLine].slice(-600));
      break;
    case 'log':
      logs.update((list) => [...list, data as LogLine].slice(-600));
      break;
    case 'traffic': {
      const event = data as TrafficEvent;
      trafficHistory.update((list) => [...list, { up: event.up, down: event.down }].slice(-180));
      totalTraffic.set({ up: event.totalUp, down: event.totalDown });
      break;
    }
    case 'running':
      running.set(data as RunningState);
      break;
    case 'coreStatus':
      coreConnected.set(Boolean((data as any)?.connected));
      break;
    case 'profiles':
      profiles.set(data as Profile[]);
      break;
    case 'providers':
      providers.set(data as Record<string, ProviderView>);
      break;
    case 'error':
      showBanner(String(data?.message ?? data), 'error');
      break;
  }
}

// setRunningState applies a core start/stop response locally.
export function setRunningState(state: RunningState) {
  running.set(state);
}

// loadBootstrap fetches initial state.
export async function loadBootstrap() {
  const state = await api.bootstrap();
  config.set(state.config);
  profiles.set(state.profiles);
  groups.set(state.groups);
  delayMap.set(state.delayMap);
  providers.set(state.providers);
  running.set(state.running);
  liveCount.set(state.liveCount);
  coreConnected.set(Boolean((state as any).coreConnected));
  trafficHistory.set(state.traffics ?? []);
}

// pushConfig merges a partial config and sends it.
export async function pushConfig(patch: ConfigPatch) {
  try {
    const result = await api.putConfig(patch);
    if (result?.config) config.set(result.config as Config);
    if (result?.warning) showBanner(result.warning, 'error');
  } catch (error) {
    showBanner(String((error as Error).message ?? error), 'error');
  }
}
