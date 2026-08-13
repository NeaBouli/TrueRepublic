/**
 * Validate mnemonic. Normalizes like the wallet service (NFKC, lowercase,
 * collapsed whitespace) so the form and the service accept exactly the same
 * phrases; wordlist membership and checksum are enforced fail-closed by
 * WalletService on import.
 */
export function validateMnemonic(mnemonic: string): {
  valid: boolean;
  error?: string;
} {
  const words = mnemonic
    .normalize('NFKC')
    .trim()
    .toLowerCase()
    .split(/\s+/)
    .filter((word) => word.length > 0);

  if (words.length !== 12 && words.length !== 24) {
    return {
      valid: false,
      error: 'Mnemonic must be 12 or 24 words',
    };
  }

  if (words.some((w) => !/^[a-z]+$/.test(w))) {
    return {
      valid: false,
      error: 'Recovery phrase words must be letters a-z',
    };
  }

  return { valid: true };
}

/**
 * Validate password
 */
export function validatePassword(password: string): {
  valid: boolean;
  error?: string;
} {
  if (password.length < 8) {
    return {
      valid: false,
      error: 'Password must be at least 8 characters',
    };
  }

  return { valid: true };
}

/**
 * Validate wallet name
 */
export function validateWalletName(name: string): {
  valid: boolean;
  error?: string;
} {
  if (!name.trim()) {
    return {
      valid: false,
      error: 'Name is required',
    };
  }

  if (name.length > 50) {
    return {
      valid: false,
      error: 'Name must be less than 50 characters',
    };
  }

  return { valid: true };
}
