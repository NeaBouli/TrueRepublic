import { describe, expect, it } from 'vitest';
import {
  validateMnemonic,
  validatePassword,
  validateWalletName,
} from './validation';

const TWELVE_WORDS =
  'abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about';

describe('validateMnemonic', () => {
  it('accepts 12 and 24 lowercase words', () => {
    expect(validateMnemonic(TWELVE_WORDS).valid).toBe(true);
    expect(validateMnemonic(`${'abandon '.repeat(23)}art`).valid).toBe(true);
  });

  it('normalizes case and surrounding whitespace like the wallet service', () => {
    expect(validateMnemonic(`  ${TWELVE_WORDS.toUpperCase()} \n`).valid).toBe(
      true
    );
    expect(
      validateMnemonic(TWELVE_WORDS.split(' ').join('   ')).valid
    ).toBe(true);
  });

  it.each([11, 13, 23])('rejects %i words', (count) => {
    const mnemonic = 'abandon '.repeat(count).trim();
    const result = validateMnemonic(mnemonic);
    expect(result.valid).toBe(false);
    expect(result.error).toContain('12 or 24 words');
  });

  it('rejects empty input', () => {
    expect(validateMnemonic('   ').valid).toBe(false);
  });

  it('rejects words that are not plain letters', () => {
    const withDigit = TWELVE_WORDS.replace('about', 'ab0ut');
    const result = validateMnemonic(withDigit);
    expect(result.valid).toBe(false);
    expect(result.error).toContain('letters a-z');
  });
});

describe('validatePassword', () => {
  it('rejects passwords shorter than 8 characters', () => {
    expect(validatePassword('short').valid).toBe(false);
    expect(validatePassword('').valid).toBe(false);
  });

  it('accepts 8 or more characters', () => {
    expect(validatePassword('12345678').valid).toBe(true);
  });
});

describe('validateWalletName', () => {
  it('rejects blank and overlong names', () => {
    expect(validateWalletName('   ').valid).toBe(false);
    expect(validateWalletName('x'.repeat(51)).valid).toBe(false);
  });

  it('accepts a normal name', () => {
    expect(validateWalletName('My Wallet').valid).toBe(true);
  });
});
