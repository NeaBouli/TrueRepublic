import { afterEach, describe, expect, it, vi } from 'vitest';
import { act, fireEvent, render, screen, waitFor } from '@testing-library/react';
import type { Identity } from '@/types/zkp';
import type { Suggestion } from '@/types/governance';
import { useIdentityStore } from '@/stores/identityStore';
import { ZKPService } from '@/services/zkp';
import { previewMockIdentityHash } from '@/services/previewIdentityHash';
import { VotingPanel } from './VotingPanel';

// GH300B3a: synthetic preview identities; never real secrets. 'c4' x32 is
// outside the BN254 scalar field, like most random preview secrets.
const OUT_OF_FIELD_SECRET = 'c4'.repeat(32);
const IN_FIELD_SECRET = '0f'.repeat(32);

const SUGGESTION: Suggestion = {
  suggestionId: 'FixtureSuggestion',
  issueId: 'FixtureIssue',
  domainId: 'FixtureDomain',
  title: 'FixtureSuggestion',
  description: '',
  creator: 'truerepublic1creator',
  avgRating: 0,
  ratingCount: 0,
  greenStones: 0,
  yellowStones: 0,
  redStones: 0,
  zone: 'unzoned',
  createdAt: '1970-01-01T00:00:00.000Z',
};

function previewIdentity(secret: string): Identity {
  return {
    secret,
    commitment: previewMockIdentityHash(secret),
    nullifier: previewMockIdentityHash(`${secret}00`),
    createdAt: 0,
  };
}

function showReadyIdentity(identity: Identity) {
  act(() => {
    useIdentityStore.setState({ identity, hasIdentity: true, status: 'ready', problem: null, legacyPresent: false });
  });
}

function renderPanel() {
  return render(<VotingPanel suggestion={SUGGESTION} domainId="FixtureDomain" issueName="FixtureIssue" />);
}

afterEach(() => {
  vi.restoreAllMocks();
  act(() => {
    useIdentityStore.setState({ identity: null, hasIdentity: false, status: 'locked', problem: null, legacyPresent: false });
  });
});

describe('VotingPanel disabled preview status (GH300B3a)', () => {
  it.each([
    ['outside the BN254 field', OUT_OF_FIELD_SECRET],
    ['inside the BN254 field', IN_FIELD_SECRET],
  ])('never derives or queries a nullifier for a preview identity %s', async (_label, secret) => {
    const external = vi.spyOn(ZKPService.prototype, 'computeExternalNullifier');
    const nullifier = vi.spyOn(ZKPService.prototype, 'computeNullifierHash');
    const query = vi.spyOn(ZKPService.prototype, 'isNullifierUsed').mockResolvedValue(true);
    showReadyIdentity(previewIdentity(secret));

    renderPanel();

    expect(await screen.findByText('ZKP preview — submission disabled')).toBeInTheDocument();
    await act(async () => {
      await Promise.resolve();
    });
    expect(external).not.toHaveBeenCalled();
    expect(nullifier).not.toHaveBeenCalled();
    expect(query).not.toHaveBeenCalled();
    expect(screen.queryByText(/Nullifier status unavailable/)).not.toBeInTheDocument();
    expect(screen.queryByText('Already Voted')).not.toBeInTheDocument();

    const control = screen.getByRole('button', { name: 'Anonymous Voting Unavailable' });
    expect(control).toBeDisabled();
    fireEvent.click(control);
    control.focus();
    expect(control).not.toHaveFocus();
    expect(screen.queryByText('Generating Anonymous Vote')).not.toBeInTheDocument();
  });

  it('still checks the canonical nullifier once submission is enabled', async () => {
    vi.spyOn(ZKPService.prototype, 'isSubmittable', 'get').mockReturnValue(true);
    const query = vi.spyOn(ZKPService.prototype, 'isNullifierUsed').mockResolvedValue(true);
    showReadyIdentity(previewIdentity(IN_FIELD_SECRET));

    renderPanel();

    await waitFor(() => expect(query).toHaveBeenCalledTimes(1));
    expect(query.mock.calls[0]?.[0]).toBe('FixtureDomain');
    expect(query.mock.calls[0]?.[1]).toMatch(/^[0-9a-f]{64}$/);
    expect(await screen.findByText('Already Voted')).toBeInTheDocument();
  });
});
