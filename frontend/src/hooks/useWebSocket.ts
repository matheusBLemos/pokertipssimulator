import { useEffect, useRef } from 'react';
import { WebSocketClient } from '../services/wsClient';
import { handleWSMessage } from '../services/wsMessageHandler';
import { WS_URL } from '../utils/constants';

export function useWebSocket(token: string | null) {
  const clientRef = useRef<WebSocketClient | null>(null);

  useEffect(() => {
    if (!token) return;

    const client = new WebSocketClient(WS_URL, token);
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
