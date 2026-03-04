import { useNavigate } from 'react-router-dom';
import { NetworkOverview } from './NetworkOverview';
import { ValidatorList } from './ValidatorList';
import { RecentBlocks } from './RecentBlocks';
import { IBCChannels } from './IBCChannels';
import { ArrowLeftIcon, GlobeAltIcon } from '@heroicons/react/24/outline';

export function NetworkExplorer() {
  const navigate = useNavigate();

  return (
    <div className="min-h-screen bg-gray-50">
      <header className="bg-white border-b border-gray-200">
        <div className="max-w-7xl mx-auto px-4 py-4">
          <button
            onClick={() => navigate('/wallet')}
            className="flex items-center gap-2 text-gray-600 hover:text-gray-900 mb-4"
          >
            <ArrowLeftIcon className="h-5 w-5" />
            Back to Wallet
          </button>

          <div className="flex items-center gap-3">
            <GlobeAltIcon className="h-8 w-8 text-primary-600" />
            <div>
              <h1 className="text-2xl font-bold">Network Explorer</h1>
              <p className="text-gray-600">
                TrueRepublic blockchain statistics and information
              </p>
            </div>
          </div>
        </div>
      </header>

      <main className="max-w-7xl mx-auto px-4 py-8">
        <div className="space-y-8">
          {/* Network Overview */}
          <section>
            <h2 className="text-lg font-semibold mb-4">Network Status</h2>
            <NetworkOverview />
          </section>

          {/* Recent Blocks */}
          <section>
            <RecentBlocks />
          </section>

          {/* Two Column Layout */}
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-8">
            <section>
              <ValidatorList />
            </section>
            <section>
              <IBCChannels />
            </section>
          </div>
        </div>
      </main>
    </div>
  );
}
