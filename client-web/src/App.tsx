import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import { CreateWallet } from '@/components/auth/CreateWallet';
import { ImportWallet } from '@/components/auth/ImportWallet';
import { UnlockWallet } from '@/components/auth/UnlockWallet';
import { WalletDashboard } from '@/components/wallet/WalletDashboard';
import { SendForm } from '@/components/wallet/SendForm';
import { DomainBrowser } from '@/components/governance/DomainBrowser';
import { IssueList } from '@/components/governance/IssueList';
import { SuggestionList } from '@/components/governance/SuggestionList';
import { PoolList } from '@/components/dex/PoolList';
import { SwapForm } from '@/components/dex/SwapForm';
import { IdentityManager } from '@/components/zkp/IdentityManager';

function App() {
  return (
    <BrowserRouter>
      <Routes>
        <Route path="/" element={<Navigate to="/unlock" replace />} />
        <Route path="/create" element={<CreateWallet />} />
        <Route path="/import" element={<ImportWallet />} />
        <Route path="/unlock" element={<UnlockWallet />} />
        <Route path="/wallet" element={<WalletDashboard />} />
        <Route path="/send" element={<SendForm />} />
        <Route path="/governance" element={<DomainBrowser />} />
        <Route path="/governance/domain/:domainId" element={<IssueList />} />
        <Route path="/governance/domain/:domainId/issue/:issueId" element={<SuggestionList />} />
        <Route path="/dex" element={<PoolList />} />
        <Route path="/dex/swap" element={<SwapForm />} />
        <Route path="/identity" element={<IdentityManager />} />
        <Route path="*" element={<Navigate to="/unlock" replace />} />
      </Routes>
    </BrowserRouter>
  );
}

export default App;
