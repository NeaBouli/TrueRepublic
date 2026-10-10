import { beforeEach, describe, expect, it, vi } from 'vitest';
import { fireEvent, render, screen, within } from '@testing-library/react';
import { MemoryRouter, Route, Routes } from 'react-router-dom';
import { useWalletStore } from '@/stores/walletStore';
import { useIdentityStore } from '@/stores/identityStore';
import { useMembershipStore } from '@/stores/membershipStore';
import { MembershipService, PREVIEW_IDENTITY_REGISTRATION_DISABLED } from '@/services/membership';
import { WalletService } from '@/services/wallet';
import { OnboardingFlow } from './OnboardingFlow';

const DOMAIN = 'GH309';
const ADDRESS = 'truerepublic1previewmember';
const NEGATIVE_SUBMIT_STATE =
  'No transaction is sent, nothing is registered on-chain and anonymous voting is not available in this client.';
// Positive promises the original registration card made; none may survive in the submit card.
const POSITIVE_ANONYMITY_PROMISES = [
  /for\s+anonymous\s+voting/i,
  /vote\s+anonymously/i,
  /submit\s+your\s+identity\s+commitment/i,
  /register\s+(zkp\s+)?identity\s+commitment/i,
  /register\s+zkp\s+identity/i,
];

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
    const heading = screen.getByRole('heading', { name: 'Identity Registration Unavailable' });
    expect(screen.getByTestId('identity-registration-disabled-notice').textContent?.replace(/\s+/g, ' ').trim()).toBe(
      `${PREVIEW_IDENTITY_REGISTRATION_DISABLED} ${NEGATIVE_SUBMIT_STATE}`
    );

    const card = heading.closest('.card');
    expect(card).not.toBeNull();
    const cardText = (card?.textContent ?? '').replace(/\s+/g, ' ');
    for (const promise of POSITIVE_ANONYMITY_PROMISES) {
      expect(cardText).not.toMatch(promise);
    }
    expect(within(card as HTMLElement).queryByText(/anonymous voting\s*$/i)).not.toBeInTheDocument();
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
