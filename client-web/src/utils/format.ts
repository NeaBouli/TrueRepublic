/**
 * Format PNYX amount (from micro to display)
 */
export function formatPnyx(amount: string | number): string {
  const num = typeof amount === 'string' ? parseInt(amount, 10) : amount;
  const pnyx = num / 1_000_000;

  return pnyx.toLocaleString('en-US', {
    minimumFractionDigits: 2,
    maximumFractionDigits: 6,
  });
}

/**
 * Parse PNYX input (from display to micro)
 */
export function parsePnyx(input: string): string {
  const num = parseFloat(input.replace(/,/g, ''));
  if (isNaN(num)) return '0';
  return Math.floor(num * 1_000_000).toString();
}

/**
 * Format address (shorten)
 */
export function formatAddress(address: string, chars: number = 8): string {
  if (address.length <= chars * 2) return address;
  return `${address.slice(0, chars)}...${address.slice(-chars)}`;
}

/**
 * Copy to clipboard
 */
export async function copyToClipboard(text: string): Promise<boolean> {
  try {
    await navigator.clipboard.writeText(text);
    return true;
  } catch {
    return false;
  }
}

/**
 * Format date
 */
export function formatDate(timestamp: number): string {
  return new Date(timestamp).toLocaleDateString('en-US', {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
  });
}
