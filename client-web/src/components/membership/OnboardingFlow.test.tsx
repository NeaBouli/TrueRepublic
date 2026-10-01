import { beforeEach, describe, expect, it, vi } from 'vitest';
import { fireEvent, render, screen } from '@testing-library/react';
import { MemoryRouter, Route, Routes } from 'react-router-dom';
import { useWalletStore } from '@/stores/walletStore';
import { useIdentityStore } from '@/stores/identityStore';
import { useMembershipStore } from '@/stores/membershipStore';
import { MembershipService, PREVIEW_IDENTITY_REGISTRATION_DISABLED } from '@/services/membership';
import { WalletService } from '@/services/wallet';
import { OnboardingFlow } from './OnboardingFlow';

const DOMAIN = 'GH309';
const ADDRESS = 'truerepublic1previewmember';

// Synthetic preview identity only; never a real secret.
const SYNTHETIC_IDENTITY = {
  secret: 'synthetic-test-secret',
  commitment: 'ab'.repeat(32),
  nullifier: 'cd'.repeat(32),
  createdAt: 0,
};

function renderSubmitStep() {
  useWalletStore.setState({
    currentWallet: { address: ADDRESS, name: 'test', createdAt: 0 },
    password: 'synthetic-password',
  });
  useIdentityStore.setState({ identity: SYNTHETIC_IDENTITY, hasIdentity: true, isInitialized: true });
  useMembershipStore.setState({
    memberships: {
      [DOMAIN]: {
        domainId: DOMAIN,
        address: ADDRESS,
        isMember: true,
        hasIdentityCommitment: false,
        inMerkleTree: false,
        step1Complete: true,
        step2Complete: true,
      },
    },
    loadMembership: vi.fn().mockResolvedValue(undefined),
  });
  render(
    <MemoryRouter initialEntries={[`/onboard/${DOMAIN}`]}>
      <Routes>
        <Route path="/onboard/:domainId" element={<OnboardingFlow />} />
      </Routes>
    </MemoryRouter>
  );
}

describe('OnboardingFlow identity registration step (issue #309)', () => {
  let registerIdentity: ReturnType<typeof vi.spyOn>;
  let getWalletForSigning: ReturnType<typeof vi.spyOn>;

  beforeEach(() => {
    registerIdentity = vi.spyOn(MembershipService.prototype, 'registerIdentity');
    getWalletForSigning = vi.spyOn(WalletService, 'getWalletForSigning');
    registerIdentity.mockClear();
    getWalletForSigning.mockClear();
  });

  it('presents registration as unavailable without promising anonymity', () => {
    renderSubmitStep();
    expect(screen.getByRole('heading', { name: 'Identity Registration Unavailable' })).toBeInTheDocument();
    expect(screen.getByTestId('identity-registration-disabled-notice')).toHaveTextContent(
      PREVIEW_IDENTITY_REGISTRATION_DISABLED
    );
    expect(screen.getByTestId('identity-registration-disabled-notice')).toHaveTextContent(
      'No transaction is sent'
    );
    expect(screen.queryByText(/anonymous voting\s*$/i)).not.toBeInTheDocument();
  });

  it('renders a disabled control that can never reach signing or the service', () => {
    renderSubmitStep();
    const control = screen.getByRole('button', { name: 'Registration Disabled in Preview' });
    expect(control).toBeDisabled();
    expect(control).toHaveAttribute('aria-describedby', 'identity-registration-disabled-reason');

    fireEvent.click(control);
    fireEvent.keyDown(control, { key: 'Enter' });
    fireEvent.keyDown(control, { key: ' ' });
    control.focus();

    expect(control).not.toHaveFocus();
    expect(getWalletForSigning).not.toHaveBeenCalled();
    expect(registerIdentity).not.toHaveBeenCalled();
    expect(screen.queryByRole('button', { name: 'Register Identity Commitment' })).not.toBeInTheDocument();
  });
});
