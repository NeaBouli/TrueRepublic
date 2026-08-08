import { describe, expect, it, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import { MemoryRouter, Route, Routes, matchRoutes } from 'react-router-dom';
import { lazy, Suspense } from 'react';
import type { ReactElement } from 'react';
import { ErrorBoundary } from '@/components/common/ErrorBoundary';

// The routing boundary is under test, not the pages. Page modules are stubbed
// with markers so no wallet, signing, store, or network code is loaded. The
// stubs are inlined because vi.mock factories are hoisted above local helpers.
vi.mock('@/components/auth/CreateWallet', () => ({
  CreateWallet: () => <div data-testid="page-create" />,
}));
vi.mock('@/components/auth/ImportWallet', () => ({
  ImportWallet: () => <div data-testid="page-import" />,
}));
vi.mock('@/components/auth/UnlockWallet', () => ({
  UnlockWallet: () => <div data-testid="page-unlock" />,
}));
vi.mock('@/components/wallet/WalletDashboard', () => ({
  WalletDashboard: () => <div data-testid="page-wallet" />,
}));
vi.mock('@/components/wallet/SendForm', () => ({
  SendForm: () => <div data-testid="page-send" />,
}));
vi.mock('@/components/governance/DomainBrowser', () => ({
  DomainBrowser: () => <div data-testid="page-governance" />,
}));
vi.mock('@/components/governance/IssueList', async () => {
  const { useParams } = await import('react-router-dom');
  return {
    IssueList: () => (
      <div data-testid="page-issues" data-params={JSON.stringify(useParams())} />
    ),
  };
});
vi.mock('@/components/governance/SuggestionList', async () => {
  const { useParams } = await import('react-router-dom');
  return {
    SuggestionList: () => (
      <div data-testid="page-suggestions" data-params={JSON.stringify(useParams())} />
    ),
  };
});
vi.mock('@/components/governance/CreateSuggestion', async () => {
  const { useParams } = await import('react-router-dom');
  return {
    CreateSuggestion: () => (
      <div data-testid="page-create-suggestion" data-params={JSON.stringify(useParams())} />
    ),
  };
});
vi.mock('@/components/dex/PoolList', () => ({
  PoolList: () => <div data-testid="page-dex" />,
}));
vi.mock('@/components/dex/SwapForm', () => ({
  SwapForm: () => <div data-testid="page-swap" />,
}));
vi.mock('@/components/dex/LPPositions', () => ({
  LPPositions: () => <div data-testid="page-positions" />,
}));
vi.mock('@/components/dex/AddLiquidity', async () => {
  const { useParams } = await import('react-router-dom');
  return {
    AddLiquidity: () => (
      <div data-testid="page-add-liquidity" data-params={JSON.stringify(useParams())} />
    ),
  };
});
vi.mock('@/components/dex/RemoveLiquidity', async () => {
  const { useParams } = await import('react-router-dom');
  return {
    RemoveLiquidity: () => (
      <div data-testid="page-remove-liquidity" data-params={JSON.stringify(useParams())} />
    ),
  };
});
vi.mock('@/components/admin/AdminDashboard', async () => {
  const { useParams } = await import('react-router-dom');
  return {
    AdminDashboard: () => (
      <div data-testid="page-admin" data-params={JSON.stringify(useParams())} />
    ),
  };
});
vi.mock('@/components/network/NetworkExplorer', () => ({
  NetworkExplorer: () => <div data-testid="page-network" />,
}));
vi.mock('@/components/zkp/IdentityManager', () => ({
  IdentityManager: () => <div data-testid="page-identity" />,
}));
vi.mock('@/components/membership/InviteHandler', async () => {
  const { useSearchParams } = await import('react-router-dom');
  return {
    InviteHandler: () => (
      <div data-testid="page-invite" data-code={useSearchParams()[0].get('code')} />
    ),
  };
});
vi.mock('@/components/membership/OnboardingFlow', async () => {
  const { useParams } = await import('react-router-dom');
  return {
    OnboardingFlow: () => (
      <div data-testid="page-onboard" data-params={JSON.stringify(useParams())} />
    ),
  };
});

import { appRoutes } from '@/routes';

const renderAt = (path: string): void => {
  render(
    <MemoryRouter initialEntries={[path]}>
      <Suspense fallback={<div role="status">Loading page…</div>}>
        <Routes>
          {appRoutes.map((route) => (
            <Route key={route.path} path={route.path} element={route.element} />
          ))}
        </Routes>
      </Suspense>
    </MemoryRouter>
  );
};

const match = (path: string) => {
  const matches = matchRoutes(appRoutes, path);
  expect(matches).not.toBeNull();
  return matches!;
};

describe('route contract', () => {
  it('defines exactly the preserved route table in order', () => {
    expect(appRoutes.map((route) => route.path)).toEqual([
      '/',
      '/create',
      '/import',
      '/unlock',
      '/wallet',
      '/send',
      '/governance',
      '/governance/domain/:domainId',
      '/governance/domain/:domainId/issue/:issueId',
      '/governance/domain/:domainId/issue/:issueId/create',
      '/dex',
      '/dex/swap',
      '/dex/positions',
      '/dex/pool/:assetDenom/add',
      '/dex/pool/:assetDenom/remove',
      '/admin/domain/:domainId',
      '/network',
      '/identity',
      '/invite',
      '/onboard/:domainId',
      '*',
    ]);
  });

  it('redirects root and the catch-all to /unlock with replace', () => {
    for (const path of ['/', '*']) {
      const route = appRoutes.find((candidate) => candidate.path === path);
      expect(route).toBeDefined();
      const element = route!.element as ReactElement<{ to: string; replace: boolean }>;
      expect(element.props.to).toBe('/unlock');
      expect(element.props.replace).toBe(true);
    }
  });
});

describe('redirects', () => {
  it('redirects the root path to /unlock', async () => {
    renderAt('/');
    expect(await screen.findByTestId('page-unlock')).toBeInTheDocument();
  });

  it('redirects unknown paths to /unlock', async () => {
    renderAt('/no-such-page');
    expect(await screen.findByTestId('page-unlock')).toBeInTheDocument();
  });

  it('matches unknown paths only through the catch-all route', () => {
    const matches = match('/definitely/not/a/route');
    expect(matches[matches.length - 1].route.path).toBe('*');
  });
});

describe('static routes', () => {
  it.each([
    ['/create', 'page-create'],
    ['/import', 'page-import'],
    ['/unlock', 'page-unlock'],
    ['/wallet', 'page-wallet'],
    ['/send', 'page-send'],
    ['/governance', 'page-governance'],
    ['/dex', 'page-dex'],
    ['/dex/swap', 'page-swap'],
    ['/dex/positions', 'page-positions'],
    ['/network', 'page-network'],
    ['/identity', 'page-identity'],
  ])('renders %s', async (path, testId) => {
    renderAt(path);
    expect(await screen.findByTestId(testId)).toBeInTheDocument();
  });

  it('matches static routes without parameters', () => {
    for (const path of ['/unlock', '/wallet', '/governance', '/dex', '/network', '/identity']) {
      const matches = match(path);
      expect(matches[matches.length - 1].params).toEqual({});
    }
  });
});

describe('parameterized routes', () => {
  it.each([
    ['/governance/domain/pnyx', 'page-issues', { domainId: 'pnyx' }],
    [
      '/governance/domain/pnyx/issue/42',
      'page-suggestions',
      { domainId: 'pnyx', issueId: '42' },
    ],
    [
      '/governance/domain/pnyx/issue/42/create',
      'page-create-suggestion',
      { domainId: 'pnyx', issueId: '42' },
    ],
    ['/dex/pool/atom/add', 'page-add-liquidity', { assetDenom: 'atom' }],
    ['/dex/pool/atom/remove', 'page-remove-liquidity', { assetDenom: 'atom' }],
    ['/admin/domain/validators', 'page-admin', { domainId: 'validators' }],
    ['/onboard/pnyx', 'page-onboard', { domainId: 'pnyx' }],
  ])('renders %s with extracted params', async (path, testId, params) => {
    renderAt(path);
    const page = await screen.findByTestId(testId);
    expect(page).toBeInTheDocument();
    expect(JSON.parse(page.getAttribute('data-params') ?? '{}')).toEqual(params);
  });

  it('matches parameterized routes through matchRoutes with the same params', () => {
    const matches = match('/governance/domain/pnyx/issue/42/create');
    expect(matches[matches.length - 1].route.path).toBe(
      '/governance/domain/:domainId/issue/:issueId/create'
    );
    expect(matches[matches.length - 1].params).toEqual({ domainId: 'pnyx', issueId: '42' });
  });
});

describe('search params', () => {
  it('preserves the invite search param for the page', async () => {
    renderAt('/invite?code=abc123');
    const page = await screen.findByTestId('page-invite');
    expect(page).toBeInTheDocument();
    expect(page.getAttribute('data-code')).toBe('abc123');
  });

  it('ignores the query string when matching', () => {
    const matches = match('/dex/swap?from=pnyx&to=atom');
    expect(matches[matches.length - 1].route.path).toBe('/dex/swap');
  });
});

describe('backslash-shaped navigation input', () => {
  it.each([
    '/\\governance',
    '\\governance',
    '/governance\\domain\\pnyx',
    '/%5Cgovernance',
  ])('never reaches a real route from %s', (path) => {
    const matches = matchRoutes(appRoutes, path);
    // Either nothing matches at all or only the catch-all redirect matches;
    // a backslash-shaped input must never resolve to a real page route.
    if (matches !== null) {
      expect(matches[matches.length - 1].route.path).toBe('*');
    }
  });

  it('renders the safe /unlock fallback for a backslash-shaped path', async () => {
    renderAt('/\\governance');
    expect(await screen.findByTestId('page-unlock')).toBeInTheDocument();
  });
});

describe('lazy route boundary', () => {
  it('renders an accessible fallback while a route chunk is loading', () => {
    const PendingPage = lazy(() => new Promise<never>(() => undefined));

    render(
      <Suspense fallback={<div role="status">Loading page…</div>}>
        <PendingPage />
      </Suspense>
    );

    expect(screen.getByRole('status')).toHaveTextContent('Loading page…');
  });

  it('passes a rejected route chunk to the application error boundary', async () => {
    const errorSpy = vi.spyOn(console, 'error').mockImplementation(() => undefined);
    const FailedPage = lazy(() => Promise.reject(new Error('chunk load failed')));

    try {
      render(
        <ErrorBoundary>
          <Suspense fallback={<div role="status">Loading page…</div>}>
            <FailedPage />
          </Suspense>
        </ErrorBoundary>
      );

      expect(await screen.findByText('Something went wrong')).toBeInTheDocument();
      expect(screen.getByText('chunk load failed')).toBeInTheDocument();
    } finally {
      errorSpy.mockRestore();
    }
  });
});
