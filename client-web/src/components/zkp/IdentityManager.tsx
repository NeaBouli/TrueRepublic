import { useNavigate } from 'react-router-dom';
import { ArrowLeftIcon } from '@heroicons/react/24/outline';
import { IdentityCustodyNotice } from './IdentityCustodyNotice';

export function IdentityManager() {
  const navigate = useNavigate();

  return (
    <div className="min-h-screen bg-gray-50">
      <header className="bg-white border-b border-gray-200">
        <div className="max-w-2xl mx-auto px-4 py-4">
          <button
            onClick={() => navigate(-1)}
            className="flex items-center gap-2 text-gray-600 hover:text-gray-900 min-h-[44px]"
          >
            <ArrowLeftIcon className="h-5 w-5" aria-hidden="true" />
            Back
          </button>
        </div>
      </header>

      <main className="max-w-2xl mx-auto px-4 py-8">
        <IdentityCustodyNotice />
      </main>
    </div>
  );
}
