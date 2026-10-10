import { Button } from '@/components/common/Button';
import { IdentityCustodyNotice } from './IdentityCustodyNotice';

interface IdentitySetupProps {
  onComplete: () => void;
}

/**
 * Former create/import screen. Until GH309C there is no way to create or
 * import a preview identity; this only shows the custody status of the
 * unlocked wallet (issue #309).
 */
export function IdentitySetup({ onComplete }: IdentitySetupProps) {
  return (
    <div className="max-w-2xl mx-auto space-y-3">
      <IdentityCustodyNotice />
      <Button variant="secondary" onClick={onComplete} className="w-full min-h-[44px]">
        Close
      </Button>
    </div>
  );
}
