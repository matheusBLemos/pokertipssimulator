import { useEffect, useRef } from 'react';
import { WebSocketClient } from '../services/wsClient';
import { handleWSMessage } from '../services/wsMessageHandler';
import { getWsUrl } from '../utils/constants';

export function useWebSocket(token: string | null) {
  const clientRef = useRef<WebSocketClient | null>(null);

  useEffect(() => {
    if (!token) return;

    const client = new WebSocketClient(getWsUrl(), token);
    clientRef.current = client;

    const unsubscribe = client.onMessage(handleWSMessage);
    client.connect();

    return () => {
      unsubscribe();
      client.disconnect();
      clientRef.current = null;
    };
  }, [token]);

  return clientRef;
}
