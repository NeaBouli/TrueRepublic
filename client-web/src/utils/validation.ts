/**
 * Validate mnemonic
 */
export function validateMnemonic(mnemonic: string): {
  valid: boolean;
  error?: string;
} {
  const words = mnemonic.trim().split(/\s+/);

  if (words.length !== 12 && words.length !== 24) {
    return {
      valid: false,
      error: 'Mnemonic must be 12 or 24 words',
    };
  }

  if (words.some((w) => !w)) {
    return {
      valid: false,
      error: 'All words must be filled',
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
