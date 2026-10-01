import type { IdentityExportFile } from '@/stores/identityStore';

/** User-triggered file download; nothing is copied to the clipboard. */
export function downloadIdentityFile(file: IdentityExportFile): void {
  const url = URL.createObjectURL(new Blob([file.content], { type: file.mimeType }));
  const link = document.createElement('a');
  link.href = url;
  link.download = file.filename;
  link.rel = 'noopener';
  document.body.appendChild(link);
  link.click();
  link.remove();
  URL.revokeObjectURL(url);
}

