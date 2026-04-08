import { useAppStore } from '../store/appStore';

export const STREETS: Record<string, string> = {
  preflop: 'Pre-flop',
  flop: 'Flop',
  turn: 'Turn',
  river: 'River',
  showdown: 'Showdown',
};

export const MAX_SEATS = 10;

export function getBaseUrl(): string {
  const addr = useAppStore.getState().serverAddress;
  if (addr) {
    const normalized = addr.replace(/\/+$/, '');
    return normalized.startsWith('http') ? normalized : `http://${normalized}`;
  }
  return '';
}

export function getWsUrl(): string {
  const addr = useAppStore.getState().serverAddress;
  if (addr) {
    const normalized = addr.replace(/\/+$/, '').replace(/^https?:\/\//, '');
    return `ws://${normalized}/ws`;
  }
  const proto = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
  return `${proto}//${window.location.host}/ws`;
}
