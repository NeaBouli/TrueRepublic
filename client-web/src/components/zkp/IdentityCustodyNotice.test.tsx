import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { act, fireEvent, render, screen } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import type { Identity } from '@/types/zkp';
import { IDENTITY_CREATION_DISABLED, useIdentityStore, type IdentityStatus } from '@/stores/identityStore';
import { ZKPService } from '@/services/zkp';
import { IdentityCustodyNotice } from './IdentityCustodyNotice';
import { IdentityManager } from './IdentityManager';

// Synthetic identity; never a real secret.
const IDENTITY: Identity = {
  secret: '5e'.repeat(32),
  commitment: '6f'.repeat(32),
  nullifier: '7a'.repeat(32),
  createdAt: 0,
};

function show(status: IdentityStatus, extra: Partial<ReturnType<typeof useIdentityStore.getState>> = {}) {
  act(() => {
    useIdentityStore.setState({
      identity: status === 'ready' ? IDENTITY : null,
      hasIdentity: status === 'ready',
      status,
      problem: null,
      legacyPresent: false,
      ...extra,
    });
  });
}

function expectInert(name: string) {
  const control = screen.getByRole('button', { name });
  expect(control).toBeDisabled();
  expect(control).toHaveAttribute('aria-describedby');
  fireEvent.click(control);
  fireEvent.keyDown(control, { key: 'Enter' });
  fireEvent.keyDown(control, { key: ' ' });
  control.focus();
  expect(control).not.toHaveFocus();
  return control;
}

describe('IdentityCustodyNotice', () => {
  const clipboard = { writeText: vi.fn() };

  beforeEach(() => {
    vi.stubGlobal('navigator', { ...globalThis.navigator, clipboard });
    clipboard.writeText.mockClear();
  });

  afterEach(() => {
    vi.unstubAllGlobals();
    vi.restoreAllMocks();
  });

  it('renders the locked state without identity data', () => {
    show('locked');
    render(<IdentityCustodyNotice />);
    expect(screen.getByRole('heading', { name: 'Preview Identity Locked' })).toBeInTheDocument();
    expect(screen.getByRole('status')).toHaveTextContent('Unlock your wallet');
  });

  it('renders absent with a disabled, unreachable creation control', () => {
    const generate = vi.spyOn(ZKPService.prototype, 'generateIdentity');
    show('absent');
    render(<IdentityCustodyNotice />);
    expect(screen.getByRole('status')).toHaveTextContent(IDENTITY_CREATION_DISABLED);
    expectInert('Identity Creation Disabled in Preview');
    const enabledEntryPoints = screen
      .queryAllByRole('button')
      .filter((button) => /creat|import/i.test(button.textContent ?? '') && !(button as HTMLButtonElement).disabled);
    expect(enabledEntryPoints).toEqual([]);
    expect(screen.queryByRole('textbox')).not.toBeInTheDocument();
    expect(generate).not.toHaveBeenCalled();
  });

  it('offers encryption and a raw file download for pending legacy data', () => {
    const exportLegacyFile = vi.fn(() => ({ filename: 'x.txt', mimeType: 'text/plain', content: 'raw' }));
    const createObjectURL = vi.fn(() => 'blob:synthetic');
    const revokeObjectURL = vi.fn();
    vi.stubGlobal('URL', { ...URL, createObjectURL, revokeObjectURL });
    const click = vi.spyOn(HTMLAnchorElement.prototype, 'click').mockImplementation(() => undefined);
    show('legacy-pending', { legacyPresent: true, exportLegacyFile });
    render(<IdentityCustodyNotice />);
    expect(screen.getByRole('heading', { name: 'Unencrypted Preview Identity Found' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'Encrypt Into This Wallet' })).toBeEnabled();
    expect(screen.getByText(/keep the file offline/)).toBeInTheDocument();
    fireEvent.click(screen.getByRole('button', { name: 'Download Unencrypted Data File' }));
    expect(exportLegacyFile).toHaveBeenCalledTimes(1);
    expect(createObjectURL).toHaveBeenCalledTimes(1);
    expect(click).toHaveBeenCalledTimes(1);
    expect(revokeObjectURL).toHaveBeenCalledWith('blob:synthetic');
    expect(clipboard.writeText).not.toHaveBeenCalled();
  });

  it.each([
    ['quarantined', 'preview-hash-mismatch', 'Preview Identity Not Used', /does not match its secret/],
    ['error', 'vault-corrupt', 'Preview Identity Unavailable', /storage on this device is damaged/],
    ['error', 'wrong-password', 'Preview Identity Unavailable', /wallet's password/],
    ['error', 'storage', 'Preview Identity Unavailable', /unavailable in this browser/],
  ] as const)('renders %s/%s with a bounded message', (status, problem, heading, message) => {
    show(status, { problem, legacyPresent: true });
    render(<IdentityCustodyNotice />);
    expect(screen.getByRole('heading', { name: heading })).toBeInTheDocument();
    expect(screen.getByRole('status')).toHaveTextContent(message);
    expect(screen.getByRole('status')).toHaveTextContent('Your wallet stays unlocked');
    expect(screen.getByRole('button', { name: 'Download Unencrypted Data File' })).toBeEnabled();
  });

  it('renders ready with file export, disabled deletion and no secret in the DOM', () => {
    const exportIdentityFile = vi.fn(() => ({ filename: 'x.json', mimeType: 'application/json', content: '{}' }));
    vi.stubGlobal('URL', { ...URL, createObjectURL: vi.fn(() => 'blob:x'), revokeObjectURL: vi.fn() });
    vi.spyOn(HTMLAnchorElement.prototype, 'click').mockImplementation(() => undefined);
    show('ready', { exportIdentityFile });
    render(<IdentityCustodyNotice />);
    expect(document.body.textContent).not.toContain(IDENTITY.secret);
    expect(document.body.textContent).not.toContain(IDENTITY.nullifier);
    expect(screen.getByTestId('identity-export-warning')).toHaveTextContent(/not canonical and not production-valid/);
    fireEvent.click(screen.getByRole('button', { name: 'Download Identity Backup File' }));
    expect(exportIdentityFile).toHaveBeenCalledTimes(1);
    expectInert('Identity Deletion Disabled in Preview');
    expect(clipboard.writeText).not.toHaveBeenCalled();
  });

  it('IdentityManager shows only the custody notice (no create, import or delete path)', () => {
    show('absent');
    render(
      <MemoryRouter>
        <IdentityManager />
      </MemoryRouter>
    );
    expect(screen.queryByRole('button', { name: 'Import Identity' })).not.toBeInTheDocument();
    expect(screen.queryByRole('button', { name: 'Confirm Delete' })).not.toBeInTheDocument();
    expect(screen.queryByRole('button', { name: /Delete Identity$/ })).not.toBeInTheDocument();
    expectInert('Identity Creation Disabled in Preview');
  });
});
