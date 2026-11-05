// Formatting helpers matching the app's TrafficShow/figures.

const units = ['B', 'KB', 'MB', 'GB', 'TB', 'PB'];

export function trafficShow(bytes: number): { value: string; unit: string } {
  let value = Math.abs(bytes);
  let unit = 0;
  while (value >= 1024 && unit < units.length - 1) {
    value /= 1024;
    unit++;
  }
  const digits = value >= 100 || unit === 0 ? 0 : value >= 10 ? 1 : 2;
  return { value: value.toFixed(digits), unit: units[unit] };
}

export function trafficText(bytes: number): string {
  const show = trafficShow(bytes);
  return `${show.value} ${show.unit}`;
}

export function speedText(bytesPerSecond: number): { value: string; unit: string } {
  const show = trafficShow(bytesPerSecond);
  return show;
}

export function memoryText(bytes: number): string {
  return trafficText(bytes);
}

// getTimeText: h:mm:ss with the hours field padded to 2 and capped at 999.
export function runTimeText(ms: number): string {
  const diff = Math.floor(ms / 1000);
  const inHours = Math.min(999, Math.floor(diff / 3600));
  if (diff / 3600 > 999) return '999:59:59';
  const inMinutes = Math.floor((diff / 60) % 60);
  const inSeconds = Math.floor(diff % 60);
  const pad = (v: number) => String(v).padStart(2, '0');
  return `${String(inHours).padStart(2, '0')}:${pad(inMinutes)}:${pad(inSeconds)}`;
}

// getLastUpdateTimeDesc: 刚刚 / N 分钟前 / N 小时前 / N 天前 / N 月前 / N 年前
export function relativeTimeText(ms: number, t: (key: string, r?: any) => string): string {
  if (!ms) return t('unknown');
  const diff = Date.now() - ms;
  const minutes = Math.floor(diff / 60000);
  if (minutes < 1) return t('justNow');
  if (minutes < 60) return t('minutesAgo', { count: minutes });
  const hours = Math.floor(minutes / 60);
  if (hours < 24) return t('hoursAgo', { count: hours });
  const days = Math.floor(hours / 24);
  if (days < 30) return t('daysAgo', { count: days });
  const months = Math.floor(days / 30);
  if (months < 12) return t('monthsAgo', { count: months });
  return t('yearsAgo', { count: Math.floor(months / 12) });
}

export function durationText(ms: number): string {
  const total = Math.floor(ms / 1000);
  const h = Math.floor(total / 3600);
  const m = Math.floor((total % 3600) / 60);
  const s = total % 60;
  const pad = (v: number) => String(v).padStart(2, '0');
  return `${pad(h)}:${pad(m)}:${pad(s)}`;
}

export function durationShortText(ms: number): string {
  if (ms < 60000) return `${Math.floor(ms / 1000)}s`;
  if (ms < 3600000) return `${Math.floor(ms / 60000)}m`;
  return `${Math.floor(ms / 3600000)}h`;
}

export function timeText(iso: string | number): string {
  const date = typeof iso === 'number' ? new Date(iso) : new Date(iso);
  if (Number.isNaN(date.getTime())) return '';
  const pad = (v: number) => String(v).padStart(2, '0');
  return `${pad(date.getHours())}:${pad(date.getMinutes())}:${pad(date.getSeconds())}`;
}

export function dateText(ms: number): string {
  if (!ms) return '';
  const date = new Date(ms);
  const pad = (v: number) => String(v).padStart(2, '0');
  return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())} ${pad(date.getHours())}:${pad(date.getMinutes())}`;
}

// delayText renders a latency value like the app: 'timeout' / 'failed' / ms.
export function delayText(value: number): string {
  if (value === -1) return 'timeout';
  if (value === -2) return 'failed';
  if (value <= 0) return '';
  return `${value}ms`;
}

// regionEmoji picks the flag for common node-name country markers.
export function regionEmoji(name: string): string {
  if (/[\u{1F1E6}-\u{1F1FF}\u{1F300}-\u{1FAFF}\u{2600}-\u{27BF}\u{FE0F}]/u.test(name.slice(0, 4))) return '';
  const table: [RegExp, string][] = [
    [/🇭🇰|香港|港|(?<![a-z])hk(?![a-z])|hong\s*kong/i, '🇭🇰'],
    [/🇹🇼|台湾|台灣|台北|(?<![a-z])tw(?![a-z])|taiwan/i, '🇹🇼'],
    [/🇯🇵|日本|东京|東京|大阪|(?<![a-z])jp(?![a-z])|japan/i, '🇯🇵'],
    [/🇸🇬|新加坡|狮城|獅城|(?<![a-z])sg(?![a-z])|singapore/i, '🇸🇬'],
    [/🇺🇸|美国|美國|洛杉矶|圣何塞|西雅图|纽约|(?<![a-z])us(?:a)?(?![a-z])|united\s*states/i, '🇺🇸'],
    [/🇰🇷|韩国|韓國|首尔|首爾|(?<![a-z])kr(?![a-z])|korea/i, '🇰🇷'],
    [/🇬🇧|英国|英國|伦敦|(?<![a-z])(uk|gb)(?![a-z])|united\s*kingdom|britain/i, '🇬🇧'],
    [/🇩🇪|德国|德國|(?<![a-z])de(?![a-z])|germany/i, '🇩🇪'],
    [/🇫🇷|法国|法國|(?<![a-z])fr(?![a-z])|france/i, '🇫🇷'],
    [/🇨🇦|加拿大|(?<![a-z])ca(?![a-z])|canada/i, '🇨🇦'],
    [/🇦🇺|澳大利亚|澳洲|(?<![a-z])au(?![a-z])|australia/i, '🇦🇺'],
    [/🇷🇺|俄罗斯|俄羅斯|(?<![a-z])ru(?![a-z])|russia/i, '🇷🇺'],
    [/🇹🇷|土耳其|(?<![a-z])tr(?![a-z])|turkey/i, '🇹🇷'],
    [/🇮🇳|印度|(?<![a-z])in(?![a-z])|india/i, '🇮🇳'],
    [/🇧🇷|巴西|(?<![a-z])br(?![a-z])|brazil/i, '🇧🇷'],
    [/🇳🇱|荷兰|荷蘭|(?<![a-z])nl(?![a-z])|netherlands/i, '🇳🇱'],
    [/🇲🇾|马来西亚|馬來西亞|(?<![a-z])my(?![a-z])|malaysia/i, '🇲🇾'],
    [/🇹🇭|泰国|泰國|(?<![a-z])th(?![a-z])|thailand/i, '🇹🇭'],
    [/🇻🇳|越南|(?<![a-z])vn(?![a-z])|vietnam/i, '🇻🇳'],
    [/🇵🇭|菲律宾|菲律賓|(?<![a-z])ph(?![a-z])|philippines/i, '🇵🇭'],
    [/🇮🇩|印尼|(?<![a-z])id(?![a-z])|indonesia/i, '🇮🇩'],
    [/🇦🇷|阿根廷|(?<![a-z])ar(?![a-z])|argentina/i, '🇦🇷'],
    [/🇲🇽|墨西哥|(?<![a-z])mx(?![a-z])|mexico/i, '🇲🇽'],
  ];
  for (const [pattern, flag] of table) {
    if (pattern.test(name)) return flag;
  }
  return '';
}
