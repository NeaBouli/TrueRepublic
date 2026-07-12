/**
 * Format PNYX amount (from micro to display)
 */
export function formatPnyx(amount: string | number): string {
  const normalized = typeof amount === 'number' ? amount.toString() : amount;
  if (!/^\d+$/.test(normalized)) return '0.00';

  const microPnyx = BigInt(normalized);
  const whole = microPnyx / 1_000_000n;
  const rawFraction = (microPnyx % 1_000_000n).toString().padStart(6, '0');
  const fraction = rawFraction.replace(/0+$/, '').padEnd(2, '0');
  const groupedWhole = whole.toString().replace(/\B(?=(\d{3})+(?!\d))/g, ',');

  return `${groupedWhole}.${fraction}`;
}

/**
 * Parse PNYX input (from display to micro)
 */
export function parsePnyx(input: string): string {
  const normalized = input.replace(/,/g, '').trim();
  if (!/^\d+(\.\d{0,6})?$/.test(normalized)) return '0';

  const [whole, fraction = ''] = normalized.split('.');
  const fractionalMicroPnyx = fraction.padEnd(6, '0');

  return (
    BigInt(whole) * 1_000_000n + BigInt(fractionalMicroPnyx || '0')
  ).toString();
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
