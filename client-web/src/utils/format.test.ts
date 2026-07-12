import { describe, expect, it } from 'vitest';
import { formatPnyx, parsePnyx } from './format';

describe('PNYX amount formatting', () => {
  it('formats micro PNYX without losing precision', () => {
    expect(formatPnyx('21000000000000')).toBe('21,000,000.00');
    expect(formatPnyx('9007199254740993')).toBe('9,007,199,254.740993');
    expect(formatPnyx('1')).toBe('0.000001');
  });

  it('parses display PNYX exactly into micro PNYX', () => {
    expect(parsePnyx('21,000,000')).toBe('21000000000000');
    expect(parsePnyx('9007199254.740993')).toBe('9007199254740993');
    expect(parsePnyx('0.000001')).toBe('1');
  });

  it('rejects invalid and over-precision inputs', () => {
    expect(parsePnyx('-1')).toBe('0');
    expect(parsePnyx('1.0000001')).toBe('0');
    expect(parsePnyx('not-an-amount')).toBe('0');
  });
});
