import { useState, useEffect } from 'react';
import { useAppStore, type ConnectionInfo as ConnInfo } from '../../store/appStore';
import { isWailsEnvironment, getConnectionInfo } from '../../services/wailsClient';
import { getBaseUrl } from '../../utils/constants';
import toast from 'react-hot-toast';

interface ConnectionInfoProps {
  roomCode: string;
  port?: number;
}

export default function ConnectionInfoPanel({ roomCode, port }: ConnectionInfoProps) {
  const storeInfo = useAppStore((s) => s.connectionInfo);
  const setConnectionInfo = useAppStore((s) => s.setConnectionInfo);
  const [info, setInfo] = useState<ConnInfo | null>(storeInfo);

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
    } else if (port) {
      fetchConnectionInfoFromApi(port).then((ci) => {
        if (ci) {
          setConnectionInfo(ci);
          setInfo(ci);
        }
      });
    }
  }, [storeInfo, port, setConnectionInfo]);

  if (!info) return null;

  const lanAddress = info.local_ip ? `${info.local_ip}:${info.port}` : null;
  const publicAddress = info.public_ip ? `${info.public_ip}:${info.port}` : null;

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

        {lanAddress && (
          <InfoRow
            label="LAN"
            value={lanAddress}
            onCopy={() => copyToClipboard(lanAddress, 'LAN address')}
          />
        )}

        {publicAddress && (
          <InfoRow
            label="Internet"
            value={publicAddress}
            onCopy={() => copyToClipboard(publicAddress, 'Public address')}
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
        Share the LAN address with friends on the same network, or the Internet address for remote play.
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
      <div>
        <span className="text-xs text-gray-500">{label}</span>
        <p className="text-white font-mono text-sm">{value}</p>
      </div>
      <button
        onClick={onCopy}
        className="px-2 py-1 text-xs bg-gray-700 hover:bg-gray-600 rounded text-gray-300 transition-colors"
      >
        Copy
      </button>
    </div>
  );
}

async function fetchConnectionInfoFromApi(port: number): Promise<ConnInfo | null> {
  try {
    const res = await fetch(`${getBaseUrl()}/api/v1/connection-info`);
    if (!res.ok) return null;
    const data = await res.json();
    return {
      local_ip: data.local_ip ?? '',
      public_ip: data.public_ip ?? '',
      port: data.port ?? port,
      upnp_ok: data.upnp_ok ?? false,
    };
  } catch {
    return null;
  }
}
