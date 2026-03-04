import { Component } from 'react';
import type { ReactNode, ErrorInfo } from 'react';
import { Card } from './Card';
import { Button } from './Button';
import { ExclamationTriangleIcon } from '@heroicons/react/24/outline';

interface Props {
  children: ReactNode;
}

interface State {
  hasError: boolean;
  error: Error | null;
}

export class ErrorBoundary extends Component<Props, State> {
  constructor(props: Props) {
    super(props);
    this.state = { hasError: false, error: null };
  }

  static getDerivedStateFromError(error: Error): State {
    return { hasError: true, error };
  }

  componentDidCatch(error: Error, errorInfo: ErrorInfo) {
    console.error('Error caught by boundary:', error, errorInfo);
  }

  handleReset = () => {
    this.setState({ hasError: false, error: null });
    window.location.href = '/wallet';
  };

  render() {
    if (this.state.hasError) {
      return (
        <div className="min-h-screen flex items-center justify-center p-4 bg-gray-50">
          <Card className="max-w-md w-full text-center">
            <ExclamationTriangleIcon className="h-16 w-16 text-red-400 mx-auto mb-4" />
            <h2 className="text-2xl font-bold mb-2">Something went wrong</h2>
            <p className="text-gray-600 mb-6">
              An unexpected error occurred. Please try again.
            </p>

            {this.state.error && (
              <div className="bg-red-50 border border-red-200 rounded-lg p-3 mb-6 text-left">
                <code className="text-xs text-red-800 break-all">
                  {this.state.error.message}
                </code>
              </div>
            )}

            <div className="space-y-3">
              <Button onClick={this.handleReset} className="w-full">
                Return to Wallet
              </Button>
              <Button
                variant="secondary"
                onClick={() => window.location.reload()}
                className="w-full"
              >
                Reload Page
              </Button>
            </div>
          </Card>
        </div>
      );
    }

    return this.props.children;
  }
}
