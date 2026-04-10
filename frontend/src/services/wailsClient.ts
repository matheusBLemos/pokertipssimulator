import type { ConnectionInfo } from '../store/appStore';

declare global {
  interface Window {
    go?: {
      main?: {
        App?: {
          StartServer: (port: number) => Promise<ConnectionInfo>;
          GetConnectionInfo: () => Promise<ConnectionInfo>;
        };
      };
    };
  }
}

function getApp() {
  return window.go?.main?.App;
}

export function isWailsEnvironment(): boolean {
  return !!getApp();
}

export async function startServer(port: number): Promise<ConnectionInfo> {
  const app = getApp();
  if (!app) {
    throw new Error('Not running in Wails environment');
  }
  return app.StartServer(port);
}

export async function getConnectionInfo(): Promise<ConnectionInfo> {
  const app = getApp();
  if (!app) {
    throw new Error('Not running in Wails environment');
  }
  return app.GetConnectionInfo();
}

