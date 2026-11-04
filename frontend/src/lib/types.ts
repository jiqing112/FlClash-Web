// Shared client types mirroring the Go backend's JSON.

export interface SubscriptionInfo {
  upload: number;
  download: number;
  total: number;
  expire: number;
}

export interface Profile {
  id: number;
  type: 'file' | 'url' | 'custom';
  label: string;
  url: string;
  ageSecretKey?: string;
  lastUpdateDate?: number;
  autoUpdate: boolean;
  autoUpdateDuration: number;
  subscriptionInfo?: SubscriptionInfo;
  selectedMap: Record<string, string>;
  unfoldSet: string[];
  extendType: 'standard' | 'script';
  matchTarget?: string;
  order?: string;
  updating?: boolean;
}

export interface ThemeProps {
  primaryColor?: number;
  primaryColors: number[];
  themeMode: 'system' | 'light' | 'dark';
  colorStyle: 'vivid' | 'tinted' | 'soft' | 'monochrome';
  pureBlack: boolean;
  sidebarBlur: boolean;
  liquidGlass: string;
  textScale: { enable: boolean; scale: number };
  systemFontFamily?: string;
}

export interface AppSettingProps {
  locale?: string | null;
  dashboardWidgets: string[];
  onlyStatisticsProxy: boolean;
  closeConnections: boolean;
  requestFeed: 'live' | 'recent';
  delayTestTimeout: number;
  tabAnimation: 'slide' | 'fade';
  floatingNavigationBar: boolean;
  autoCheckUpdate: boolean;
  autoRun: boolean;
  openLogs: boolean;
  sidebarExpanded: boolean;
  minimizeOnExit: boolean;
  restoreStrategy: string;
  checkCertificate: boolean;
  userAgents: string[];
  hideIp: boolean;
  editorLineWrap: boolean;
  resourceUpdateInterval?: number;
}

export interface NetworkProps {
  systemProxy: boolean;
  bypassDomain: string[];
  bypassPrivateRoute: boolean;
  autoSetSystemDns: boolean;
  appendSystemDns: boolean;
  authentication: { enable: boolean; username: string; password: string };
}

export interface ProxiesStyleProps {
  type: 'tab' | 'list';
  sortType: 'none' | 'delay' | 'name';
  layout: 'loose' | 'standard' | 'tight';
  iconStyle: 'filled' | 'plain' | 'hidden';
  cardType: 'loose' | 'standard' | 'compact';
  hideTimeoutProxies: boolean;
}

export interface TunProps {
  enable: boolean;
  device: string;
  stack: string;
  'route-address': string[];
  mtu: number;
  'dns-hijack': string[];
  'auto-route': boolean;
  'auto-detect-interface': boolean;
}

export interface DnsProps {
  enable: boolean;
  listen: string;
  'use-hosts': boolean;
  'use-system-hosts': boolean;
  'respect-rules': boolean;
  ipv6: boolean;
  'cache-algorithm': string;
  'cache-max-size': number;
  'default-nameserver': string[];
  'enhanced-mode': string;
  'fake-ip-range': string;
  'fake-ip-filter': string[];
  'nameserver-policy': Record<string, string>;
  nameserver: string[];
  fallback: string[];
  'proxy-server-nameserver': string[];
  'direct-nameserver': string[];
}

export interface PatchClashConfig {
  'mixed-port': number;
  'socks-port': number;
  port: number;
  'redir-port': number;
  'tproxy-port': number;
  mode: 'rule' | 'global' | 'direct';
  'allow-lan': boolean;
  'log-level': string;
  ipv6: boolean;
  'find-process-mode': string;
  'interface-name-mode': string;
  'interface-name': string;
  'keep-alive-interval': number;
  'keep-alive-idle': number;
  'disable-keep-alive': boolean;
  'routing-mark': number;
  'unified-delay': boolean;
  'tcp-concurrent': boolean;
  tun: TunProps;
  dns: DnsProps;
  'geox-url': Record<string, string>;
  'geodata-loader': string;
  'global-ua'?: string;
  'external-controller': string;
  secret: string;
  hosts: Record<string, string>;
  'geo-auto-update': boolean;
  'geo-update-interval': number;
  'test-url': string;
}

export interface Config {
  currentProfileId?: number | null;
  appSettingProps: AppSettingProps;
  networkProps: NetworkProps;
  vpnProps: { enable: boolean; systemProxy: boolean; ipv6: boolean; allowBypass: boolean; dnsHijacking: boolean };
  themeProps: ThemeProps;
  proxiesStyleProps: ProxiesStyleProps;
  patchClashConfig: PatchClashConfig;
}

export interface RunningState {
  running: boolean;
  startTime: number;
  systemProxy: boolean;
  port: number;
  mode: string;
}

export interface Proxy {
  name: string;
  type: string;
  'provider-name'?: string;
  now?: string;
  all?: string[];
  hidden?: boolean;
  testUrl?: string;
  icon?: string;
}

export interface Group {
  name: string;
  type: string;
  all: Proxy[];
  now?: string;
  hidden?: boolean;
  testUrl?: string;
  icon?: string;
}

export interface Metadata {
  network: string;
  type: string;
  sourceIP: string;
  destinationIP: string;
  sourcePort: string;
  destinationPort: string;
  host: string;
  process: string;
  processPath: string;
  remoteDestination: string;
}

export interface RequestView {
  id: string;
  metadata: Metadata;
  upload: number;
  download: number;
  start: string;
  chains: string[];
  rule: string;
  rulePayload: string;
  uploadSpeed?: number;
  downloadSpeed?: number;
}

export interface DnsQueryLine {
  domain: string;
  type: string;
  initiator: string;
  upstream?: string;
  cached?: boolean;
  answers: string[];
  rcode?: string;
  error?: string;
  delay: number;
  time: string;
}

export interface LogLine {
  LogLevel: string;
  Payload: string;
}

export interface ProviderView {
  name: string;
  type: string;
  'vehicle-type': string;
  count: number;
  path: string;
  'update-at': string;
  'subscription-info'?: SubscriptionInfo | null;
}

export interface DelayEvent {
  url: string;
  name: string;
  value: number;
}

export interface TrafficEvent {
  up: number;
  down: number;
  totalUp: number;
  totalDown: number;
}

export interface BootstrapState {
  config: Config;
  profiles: Profile[];
  running: RunningState;
  groups: Group[];
  delayMap: Record<string, Record<string, number>>;
  providers: Record<string, ProviderView>;
  liveCount: number;
  traffics: { up: number; down: number }[];
}

export interface ConfigPatch {
  appSettingProps?: Partial<AppSettingProps>;
  themeProps?: Partial<ThemeProps>;
  networkProps?: Partial<NetworkProps>;
  vpnProps?: Partial<{ enable: boolean; systemProxy: boolean; ipv6: boolean; allowBypass: boolean; dnsHijacking: boolean }>;
  proxiesStyleProps?: Partial<ProxiesStyleProps>;
  patchClashConfig?: Partial<PatchClashConfig>;
  currentProfileId?: number;
}
