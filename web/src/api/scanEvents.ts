import { apiClient } from './client';

export interface ScanProgressEvent {
  libraryId: number;
  scanRunId: number;
  filesTotal: number;
  filesDone: number;
  status: string;
  message?: string;
}

export function subscribeScanProgress(
  libraryId: number,
  onProgress: (event: ScanProgressEvent) => void,
): () => void {
  const url = `${apiClient.baseURL()}/events/stream`;
  const source = new EventSource(url);

  source.addEventListener('scan.progress', (ev) => {
    try {
      const data = JSON.parse((ev as MessageEvent<string>).data) as ScanProgressEvent;
      if (data.libraryId === libraryId) {
        onProgress(data);
      }
    } catch {
      // ignore malformed payloads
    }
  });

  return () => source.close();
}
