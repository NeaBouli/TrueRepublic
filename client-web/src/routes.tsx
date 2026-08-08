import { lazy } from 'react';
import { Navigate } from 'react-router-dom';
import type { RouteObject } from 'react-router-dom';

const CreateWallet = lazy(() =>
  import('@/components/auth/CreateWallet').then(({ CreateWallet }) => ({ default: CreateWallet }))
);
const ImportWallet = lazy(() =>
  import('@/components/auth/ImportWallet').then(({ ImportWallet }) => ({ default: ImportWallet }))
);
const UnlockWallet = lazy(() =>
  import('@/components/auth/UnlockWallet').then(({ UnlockWallet }) => ({ default: UnlockWallet }))
);
const WalletDashboard = lazy(() =>
  import('@/components/wallet/WalletDashboard').then(({ WalletDashboard }) => ({
    default: WalletDashboard,
  }))
);
const SendForm = lazy(() =>
  import('@/components/wallet/SendForm').then(({ SendForm }) => ({ default: SendForm }))
);
const DomainBrowser = lazy(() =>
  import('@/components/governance/DomainBrowser').then(({ DomainBrowser }) => ({
    default: DomainBrowser,
  }))
);
const IssueList = lazy(() =>
  import('@/components/governance/IssueList').then(({ IssueList }) => ({ default: IssueList }))
);
const SuggestionList = lazy(() =>
  import('@/components/governance/SuggestionList').then(({ SuggestionList }) => ({
    default: SuggestionList,
  }))
);
const CreateSuggestion = lazy(() =>
  import('@/components/governance/CreateSuggestion').then(({ CreateSuggestion }) => ({
    default: CreateSuggestion,
  }))
);
const PoolList = lazy(() =>
  import('@/components/dex/PoolList').then(({ PoolList }) => ({ default: PoolList }))
);
const SwapForm = lazy(() =>
  import('@/components/dex/SwapForm').then(({ SwapForm }) => ({ default: SwapForm }))
);
const LPPositions = lazy(() =>
  import('@/components/dex/LPPositions').then(({ LPPositions }) => ({ default: LPPositions }))
);
const AddLiquidity = lazy(() =>
  import('@/components/dex/AddLiquidity').then(({ AddLiquidity }) => ({ default: AddLiquidity }))
);
const RemoveLiquidity = lazy(() =>
  import('@/components/dex/RemoveLiquidity').then(({ RemoveLiquidity }) => ({
    default: RemoveLiquidity,
  }))
);
const AdminDashboard = lazy(() =>
  import('@/components/admin/AdminDashboard').then(({ AdminDashboard }) => ({
    default: AdminDashboard,
  }))
);
const NetworkExplorer = lazy(() =>
  import('@/components/network/NetworkExplorer').then(({ NetworkExplorer }) => ({
    default: NetworkExplorer,
  }))
);
const IdentityManager = lazy(() =>
  import('@/components/zkp/IdentityManager').then(({ IdentityManager }) => ({
    default: IdentityManager,
  }))
);
const InviteHandler = lazy(() =>
  import('@/components/membership/InviteHandler').then(({ InviteHandler }) => ({
    default: InviteHandler,
  }))
);
const OnboardingFlow = lazy(() =>
  import('@/components/membership/OnboardingFlow').then(({ OnboardingFlow }) => ({
    default: OnboardingFlow,
  }))
);

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
