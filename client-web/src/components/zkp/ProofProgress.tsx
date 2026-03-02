import type { ProofGenerationStatus } from '@/types/zkp';
import { CheckCircleIcon, XCircleIcon } from '@heroicons/react/24/outline';

interface ProofProgressProps {
  status: ProofGenerationStatus;
}

const STEP_CONFIG = [
  {
    key: 'loading_wasm',
    label: 'Loading ZKP Library',
    description: 'Initializing zero-knowledge proof system',
  },
  {
    key: 'fetching_proof',
    label: 'Fetching Membership Proof',
    description: 'Getting your proof of domain membership',
  },
  {
    key: 'generating',
    label: 'Generating Anonymous Proof',
    description:
      'Creating zero-knowledge proof (this may take 20-30 seconds)',
  },
] as const;

const STEP_KEYS = STEP_CONFIG.map((s) => s.key);

type StepStatus = 'complete' | 'active' | 'pending' | 'error';

function getStepStatus(
  currentStep: ProofGenerationStatus['step'],
  stepKey: string
): StepStatus {
  if (currentStep === 'error') return 'error';
  if (currentStep === 'complete') return 'complete';

  const currentIndex = STEP_KEYS.indexOf(
    currentStep as (typeof STEP_KEYS)[number]
  );
  const stepIndex = STEP_KEYS.indexOf(stepKey as (typeof STEP_KEYS)[number]);

  if (stepIndex < currentIndex) return 'complete';
  if (stepIndex === currentIndex) return 'active';
  return 'pending';
}

function StepIcon({ stepStatus }: { stepStatus: StepStatus }) {
  switch (stepStatus) {
    case 'complete':
      return <CheckCircleIcon className="h-6 w-6 text-green-600" />;
    case 'active':
      return (
        <div className="h-6 w-6 border-4 border-primary-600 border-t-transparent rounded-full animate-spin" />
      );
    case 'error':
      return <XCircleIcon className="h-6 w-6 text-red-600" />;
    default:
      return (
        <div className="h-6 w-6 border-2 border-gray-300 rounded-full" />
      );
  }
}

function stepTextColor(stepStatus: StepStatus): string {
  switch (stepStatus) {
    case 'active':
      return 'text-primary-900';
    case 'complete':
      return 'text-green-900';
    case 'error':
      return 'text-red-900';
    default:
      return 'text-gray-500';
  }
}

export function ProofProgress({ status }: ProofProgressProps) {
  return (
    <div className="space-y-6">
      {/* Progress Bar */}
      <div className="space-y-2">
        <div className="flex items-center justify-between text-sm">
          <span className="font-medium text-gray-700">{status.message}</span>
          <span className="text-gray-500">{status.progress}%</span>
        </div>
        <div className="w-full bg-gray-200 rounded-full h-2">
          <div
            className={`h-2 rounded-full transition-all duration-500 ${
              status.step === 'error' ? 'bg-red-600' : 'bg-primary-600'
            }`}
            style={{ width: `${status.progress}%` }}
          />
        </div>
      </div>

      {/* Steps */}
      <div className="space-y-4">
        {STEP_CONFIG.map((step) => {
          const ss = getStepStatus(status.step, step.key);

          return (
            <div key={step.key} className="flex items-start gap-3">
              <div className="flex-shrink-0 mt-0.5">
                <StepIcon stepStatus={ss} />
              </div>
              <div className="flex-1">
                <div className={`font-medium ${stepTextColor(ss)}`}>
                  {step.label}
                </div>
                <div className="text-sm text-gray-600 mt-0.5">
                  {step.description}
                </div>
              </div>
            </div>
          );
        })}
      </div>

      {/* Error */}
      {status.step === 'error' && status.error && (
        <div className="bg-red-50 border border-red-200 rounded-lg p-4">
          <div className="flex items-start gap-2">
            <XCircleIcon className="h-5 w-5 text-red-600 flex-shrink-0 mt-0.5" />
            <div>
              <div className="font-medium text-red-900">
                Proof Generation Failed
              </div>
              <div className="text-sm text-red-800 mt-1">{status.error}</div>
            </div>
          </div>
        </div>
      )}

      {/* Success */}
      {status.step === 'complete' && (
        <div className="bg-green-50 border border-green-200 rounded-lg p-4">
          <div className="flex items-start gap-2">
            <CheckCircleIcon className="h-5 w-5 text-green-600 flex-shrink-0 mt-0.5" />
            <div>
              <div className="font-medium text-green-900">
                Proof Generated!
              </div>
              <div className="text-sm text-green-800 mt-1">
                Your anonymous vote is ready to submit
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
