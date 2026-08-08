import { Suspense } from 'react';
import { BrowserRouter, Routes, Route } from 'react-router-dom';
import { ErrorBoundary } from '@/components/common/ErrorBoundary';
import { ToastContainer } from '@/components/common/Toast';
import { MobileNav } from '@/components/common/MobileNav';
import { appRoutes } from '@/routes';

function App() {
  return (
    <ErrorBoundary>
      <BrowserRouter>
        <ToastContainer />
        <MobileNav />
        <Suspense
          fallback={
            <div
              className="min-h-screen flex items-center justify-center text-gray-600"
              role="status"
              aria-live="polite"
            >
              Loading page…
            </div>
          }
        >
          <Routes>
            {appRoutes.map((route) => (
              <Route key={route.path} path={route.path} element={route.element} />
            ))}
          </Routes>
        </Suspense>
      </BrowserRouter>
    </ErrorBoundary>
  );
}

export default App;
