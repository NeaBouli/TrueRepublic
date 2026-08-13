import { beforeEach, describe, expect, it, vi } from 'vitest';
import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import { toBech32 } from '@cosmjs/encoding';
import { useWalletStore } from '@/stores/walletStore';
import { SendForm } from './SendForm';

const RECIPIENT = toBech32(
  'truerepublic',
  Uint8Array.from({ length: 20 }, (_, index) => index + 1)
);

describe('SendForm amount validation', () => {
  const sendTokens = vi.fn();

  beforeEach(() => {
    sendTokens.mockReset();
    sendTokens.mockResolvedValue({ hash: 'AB'.repeat(32), height: 1, success: true });
    useWalletStore.setState({
      balances: [{ denom: 'upnyx', amount: '10000000' }],
      isLoading: false,
      sendTokens,
    });
  });

  function renderForm() {
    render(
      <MemoryRouter>
        <SendForm />
      </MemoryRouter>
    );
    fireEvent.change(screen.getByLabelText('Recipient Address'), {
      target: { value: RECIPIENT },
    });
  }

  it.each(['0', '0.000000', '1.0000001'])('rejects unsafe amount %s before signing', (amount) => {
    renderForm();
    fireEvent.change(screen.getByRole('spinbutton'), { target: { value: amount } });

    fireEvent.click(screen.getByRole('button', { name: 'Send Transaction' }));

    expect(screen.getByText(/positive amount with up to 6 decimals/i)).toBeInTheDocument();
    expect(sendTokens).not.toHaveBeenCalled();
  });

  it('submits the exact strictly parsed base amount', async () => {
    renderForm();
    fireEvent.change(screen.getByRole('spinbutton'), { target: { value: '1.000001' } });

    fireEvent.click(screen.getByRole('button', { name: 'Send Transaction' }));

    await waitFor(() =>
      expect(sendTokens).toHaveBeenCalledWith({
        to: RECIPIENT,
        amount: '1000001',
        denom: 'upnyx',
        memo: undefined,
      })
    );
  });
});
