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
      <Routes>
        {appRoutes.map((route) => (
          <Route key={route.path} path={route.path} element={route.element} />
        ))}
      </Routes>
    </BrowserRouter>
    </ErrorBoundary>
  );
}

export default App;
