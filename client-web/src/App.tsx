import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import { CreateWallet } from '@/components/auth/CreateWallet';
import { ImportWallet } from '@/components/auth/ImportWallet';
import { UnlockWallet } from '@/components/auth/UnlockWallet';
import { WalletDashboard } from '@/components/wallet/WalletDashboard';

function App() {
  return (
    <BrowserRouter>
      <Routes>
        <Route path="/" element={<Navigate to="/unlock" replace />} />
        <Route path="/create" element={<CreateWallet />} />
        <Route path="/import" element={<ImportWallet />} />
        <Route path="/unlock" element={<UnlockWallet />} />
        <Route path="/wallet" element={<WalletDashboard />} />
        <Route path="*" element={<Navigate to="/unlock" replace />} />
      </Routes>
    </BrowserRouter>
  );
}

export default App;
