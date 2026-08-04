import { Navigate } from 'react-router-dom';
import type { RouteObject } from 'react-router-dom';
import { CreateWallet } from '@/components/auth/CreateWallet';
import { ImportWallet } from '@/components/auth/ImportWallet';
import { UnlockWallet } from '@/components/auth/UnlockWallet';
import { WalletDashboard } from '@/components/wallet/WalletDashboard';
import { SendForm } from '@/components/wallet/SendForm';
import { DomainBrowser } from '@/components/governance/DomainBrowser';
import { IssueList } from '@/components/governance/IssueList';
import { SuggestionList } from '@/components/governance/SuggestionList';
import { CreateSuggestion } from '@/components/governance/CreateSuggestion';
import { PoolList } from '@/components/dex/PoolList';
import { SwapForm } from '@/components/dex/SwapForm';
import { LPPositions } from '@/components/dex/LPPositions';
import { AddLiquidity } from '@/components/dex/AddLiquidity';
import { RemoveLiquidity } from '@/components/dex/RemoveLiquidity';
import { AdminDashboard } from '@/components/admin/AdminDashboard';
import { NetworkExplorer } from '@/components/network/NetworkExplorer';
import { IdentityManager } from '@/components/zkp/IdentityManager';
import { InviteHandler } from '@/components/membership/InviteHandler';
import { OnboardingFlow } from '@/components/membership/OnboardingFlow';

/**
 * Single source of truth for the application route table. Keep the order
 * exactly as rendered in App.tsx; the routing tests in routes.test.tsx pin
 * this contract (paths, order, params, and redirect targets).
 */
export const appRoutes: RouteObject[] = [
  { path: '/', element: <Navigate to="/unlock" replace /> },
  { path: '/create', element: <CreateWallet /> },
  { path: '/import', element: <ImportWallet /> },
  { path: '/unlock', element: <UnlockWallet /> },
  { path: '/wallet', element: <WalletDashboard /> },
  { path: '/send', element: <SendForm /> },
  { path: '/governance', element: <DomainBrowser /> },
  { path: '/governance/domain/:domainId', element: <IssueList /> },
  { path: '/governance/domain/:domainId/issue/:issueId', element: <SuggestionList /> },
  { path: '/governance/domain/:domainId/issue/:issueId/create', element: <CreateSuggestion /> },
  { path: '/dex', element: <PoolList /> },
  { path: '/dex/swap', element: <SwapForm /> },
  { path: '/dex/positions', element: <LPPositions /> },
  { path: '/dex/pool/:assetDenom/add', element: <AddLiquidity /> },
  { path: '/dex/pool/:assetDenom/remove', element: <RemoveLiquidity /> },
  { path: '/admin/domain/:domainId', element: <AdminDashboard /> },
  { path: '/network', element: <NetworkExplorer /> },
  { path: '/identity', element: <IdentityManager /> },
  { path: '/invite', element: <InviteHandler /> },
  { path: '/onboard/:domainId', element: <OnboardingFlow /> },
  { path: '*', element: <Navigate to="/unlock" replace /> },
];
