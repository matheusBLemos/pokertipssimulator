import { useState, useEffect, useRef } from 'react';
import { useAppStore, type ConnectionInfo as ConnInfo } from '../../store/appStore';
import { isWailsEnvironment, getConnectionInfo } from '../../services/wailsClient';
import { getBaseUrl } from '../../utils/constants';
import toast from 'react-hot-toast';

interface ConnectionInfoProps {
  roomCode: string;
}

function buildJoinLink(baseUrl: string, roomCode: string): string {
  return `${baseUrl.replace(/\/+$/, '')}/#/game?code=${roomCode}`;
}

export default function ConnectionInfoPanel({ roomCode }: ConnectionInfoProps) {
  const storeInfo = useAppStore((s) => s.connectionInfo);
  const setConnectionInfo = useAppStore((s) => s.setConnectionInfo);
  const [info, setInfo] = useState<ConnInfo | null>(storeInfo);
  const autoCopied = useRef(false);

  useEffect(() => {
    if (storeInfo) {
      setInfo(storeInfo);
      return;
    }

    if (isWailsEnvironment()) {
      getConnectionInfo().then((ci) => {
        setConnectionInfo(ci);
        setInfo(ci);
      }).catch(() => {});
    } else {
      fetchConnectionInfoFromApi().then((ci) => {
        if (ci) {
          setConnectionInfo(ci);
          setInfo(ci);
        }
      });
    }
  }, [storeInfo, setConnectionInfo]);

  // Auto-copy the best URL when info first becomes available
  useEffect(() => {
    if (!info || autoCopied.current) return;
    autoCopied.current = true;

    const lanUrl = info.local_url ? buildJoinLink(info.local_url, roomCode) : null;
    const publicUrl = info.public_url ? buildJoinLink(info.public_url, roomCode) : null;

    const url = lanUrl || publicUrl;
    if (url) {
      navigator.clipboard.writeText(url).then(() => {
        toast.success('Join link copied to clipboard!');
      }).catch(() => {});
    }
  }, [info, roomCode]);

  if (!info) return null;

  const lanUrl = info.local_url ? buildJoinLink(info.local_url, roomCode) : null;
  const publicUrl = info.public_url ? buildJoinLink(info.public_url, roomCode) : null;

  const copyToClipboard = (text: string, label: string) => {
    navigator.clipboard.writeText(text);
    toast.success(`${label} copied!`);
  };

  return (
    <div className="bg-gray-900 rounded-xl p-4 space-y-3">
      <h3 className="text-sm font-medium text-gray-400">
        Connection Info
      </h3>

      <div className="space-y-2">
        <InfoRow
          label="Room Code"
          value={roomCode}
          onCopy={() => copyToClipboard(roomCode, 'Room code')}
        />

        {lanUrl && (
          <InfoRow
            label="LAN Link"
            value={lanUrl}
            onCopy={() => copyToClipboard(lanUrl, 'LAN link')}
          />
        )}

        {publicUrl && (
          <InfoRow
            label="Internet Link"
            value={publicUrl}
            onCopy={() => copyToClipboard(publicUrl, 'Internet link')}
          />
        )}

        <div className="flex items-center gap-2 text-xs">
          <span
            className={`w-2 h-2 rounded-full ${
              info.upnp_ok ? 'bg-green-500' : 'bg-yellow-500'
            }`}
          />
          <span className="text-gray-500">
            {info.upnp_ok
              ? 'Port mapped automatically (UPnP)'
              : 'Manual port forwarding may be required for internet play'}
          </span>
        </div>
      </div>

      <p className="text-xs text-gray-600">
        Share the LAN link with friends on the same network, or the Internet link for remote play.
      </p>
    </div>
  );
}

function InfoRow({
  label,
  value,
  onCopy,
}: {
  label: string;
  value: string;
  onCopy: () => void;
}) {
  return (
    <div className="flex items-center justify-between bg-gray-800/50 rounded-lg px-3 py-2">
      <div className="min-w-0 flex-1 mr-2">
        <span className="text-xs text-gray-500">{label}</span>
        <p className="text-white font-mono text-sm truncate">{value}</p>
      </div>
      <button
        onClick={onCopy}
        className="shrink-0 px-2 py-1 text-xs bg-gray-700 hover:bg-gray-600 rounded text-gray-300 transition-colors"
      >
        Copy
      </button>
    </div>
  );
}

async function fetchConnectionInfoFromApi(): Promise<ConnInfo | null> {
  try {
    const res = await fetch(`${getBaseUrl()}/api/v1/connection-info`);
    if (!res.ok) return null;
    const data = await res.json();
    return {
      local_ip: data.local_ip ?? '',
      public_ip: data.public_ip ?? '',
      port: data.port ?? 0,
      upnp_ok: data.upnp_ok ?? false,
      local_url: data.local_url ?? '',
      public_url: data.public_url ?? '',
    };
  } catch {
    return null;
  }
}
