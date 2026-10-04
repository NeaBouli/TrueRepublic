// Test-only entry: actual components/styles, synthetic state, no keys or network.
// This HTML is not a production Vite build entry and is not copied by Docker.
import { createRoot } from 'react-dom/client';
import { MemoryRouter } from 'react-router-dom';
import { CreateWallet } from '../../src/components/auth/CreateWallet';
import { WalletDashboard } from '../../src/components/wallet/WalletDashboard';
import { MobileNav } from '../../src/components/common/MobileNav';
import { useWalletStore } from '../../src/stores/walletStore';
import '../../src/index.css';

const mode = new URLSearchParams(window.location.search).get('mode');
const wallet = { name: 'Layout fixture', address: `truerepublic1${'q'.repeat(38)}`, createdAt: 1 };
useWalletStore.setState({
  currentWallet: mode === 'wallet' ? wallet : null,
  isLocked: mode !== 'wallet',
  historyAddress: wallet.address,
  historyStatus: 'ready',
  refreshBalance: async () => {},
  loadHistoryPage: async () => {},
  createWallet: async () => {
    useWalletStore.setState({ currentWallet: wallet, isLocked: false });
    // Deliberately invalid repeated-word fixture, never a usable recovery key.
    return { ...wallet, mnemonic: Array(24).fill('sentence').join(' ') };
  },
});

createRoot(document.getElementById('root')!).render(
  <MemoryRouter initialEntries={[mode === 'wallet' ? '/wallet' : '/create']}>
    <MobileNav />
    {mode === 'wallet' ? <WalletDashboard /> : <CreateWallet />}
  </MemoryRouter>
);
