import { useState, useEffect, useCallback } from 'react';
import { useNavigate } from 'react-router-dom';
import { SigningStargateClient, GasPrice } from '@cosmjs/stargate';
import { useWalletStore } from '@/stores/walletStore';
import { useDEXStore } from '@/stores/dexStore';
import { WalletService } from '@/services/wallet';
import { Card } from '@/components/common/Card';
import { Button } from '@/components/common/Button';
import { AssetSelector } from './AssetSelector';
import { formatPnyx, parsePnyx } from '@/utils/format';
import { DEFAULT_CHAIN } from '@/config/chains';
import {
  ArrowLeftIcon,
  ArrowsUpDownIcon,
  CheckCircleIcon,
  ExclamationTriangleIcon,
} from '@heroicons/react/24/outline';

export function SwapForm() {
  const navigate = useNavigate();
  const { currentWallet, password, balances, refreshBalance } =
    useWalletStore();
  const { swapEstimate, estimateSwap, clearEstimate, assets } = useDEXStore();

  const [inputDenom, setInputDenom] = useState<string | null>('pnyx');
  const [outputDenom, setOutputDenom] = useState<string | null>(null);
  const [inputAmount, setInputAmount] = useState('');
  const [slippage, setSlippage] = useState('1');
  const [isSwapping, setIsSwapping] = useState(false);
  const [txHash, setTxHash] = useState('');
  const [error, setError] = useState('');

  const inputBalance = balances.find((b) => b.denom === inputDenom);
  const inputAsset = assets.find((a) => a.ibc_denom === inputDenom);
  const outputAsset = assets.find((a) => a.ibc_denom === outputDenom);

  // Auto-estimate when inputs change
  const doEstimate = useCallback(() => {
    if (
      inputDenom &&
      outputDenom &&
      inputAmount &&
      parseFloat(inputAmount) > 0
    ) {
      const amountMicro = parsePnyx(inputAmount);
      estimateSwap(inputDenom, amountMicro, outputDenom);
    } else {
      clearEstimate();
    }
  }, [inputDenom, outputDenom, inputAmount, estimateSwap, clearEstimate]);

  useEffect(() => {
    doEstimate();
  }, [doEstimate]);

  const handleFlipAssets = () => {
    setInputDenom(outputDenom);
    setOutputDenom(inputDenom);
    setInputAmount('');
    clearEstimate();
  };

  const handleSetMax = () => {
    if (inputBalance) {
      const reserved = BigInt(10000);
      const max = BigInt(inputBalance.amount) - reserved;
      if (max > 0n) {
        setInputAmount(formatPnyx(max.toString()));
      }
    }
  };

  const calculateMinOutput = (): string => {
    if (!swapEstimate) return '0';
    const slippagePercent = parseFloat(slippage) || 1;
    const output = BigInt(swapEstimate.expected_output);
    const slippageAmount =
      (output * BigInt(Math.floor(slippagePercent * 100))) / 10000n;
    return (output - slippageAmount).toString();
  };

  const handleExecuteSwap = async () => {
    if (
      !currentWallet ||
      !password ||
      !inputDenom ||
      !outputDenom ||
      !swapEstimate
    ) {
      return;
    }

    try {
      setIsSwapping(true);
      setError('');

      const signingWallet = await WalletService.getWalletForSigning(
        currentWallet.address,
        password
      );

      const [account] = await signingWallet.getAccounts();

      const client = await SigningStargateClient.connectWithSigner(
        DEFAULT_CHAIN.rpc,
        signingWallet,
        { gasPrice: GasPrice.fromString(DEFAULT_CHAIN.gasPrice) }
      );

      const inputAmtMicro = parsePnyx(inputAmount);
      const minOutput = calculateMinOutput();

      const msg = {
        typeUrl: '/dex.MsgSwapExact',
        value: {
          sender: account.address,
          inputDenom,
          inputAmt: parseInt(inputAmtMicro, 10),
          outputDenom,
          minOutput: parseInt(minOutput, 10),
        },
      };

      const gasEstimate = await client
        .simulate(account.address, [msg], '')
        .catch(() => 200000);
      const gas = Math.ceil(
        (typeof gasEstimate === 'number' ? gasEstimate : 200000) * 1.3
      );

      const result = await client.signAndBroadcast(
        account.address,
        [msg],
        {
          amount: [
            { denom: DEFAULT_CHAIN.coinMinimalDenom, amount: '5000' },
          ],
          gas: gas.toString(),
        },
        ''
      );

      client.disconnect();

      if (result.code !== 0) {
        throw new Error(result.rawLog || 'Swap failed');
      }

      setTxHash(result.transactionHash);
      await refreshBalance();
    } catch (err: unknown) {
      const message =
        err instanceof Error ? err.message : 'Swap failed';
      setError(message);
    } finally {
      setIsSwapping(false);
    }
  };

  if (txHash) {
    return (
      <div className="min-h-screen flex items-center justify-center p-4">
        <Card className="max-w-md w-full text-center">
          <CheckCircleIcon className="h-16 w-16 text-green-600 mx-auto mb-4" />
          <h2 className="text-2xl font-bold mb-2">Swap Successful!</h2>
          <p className="text-gray-600 mb-6">Your swap has been executed</p>

          <div className="bg-gray-50 rounded-lg p-4 mb-6">
            <div className="text-xs text-gray-600 mb-1">Transaction Hash</div>
            <code className="text-xs font-mono break-all">{txHash}</code>
          </div>

          <div className="space-y-3">
            <Button onClick={() => navigate('/dex')} className="w-full">
              Back to DEX
            </Button>
            <Button
              variant="secondary"
              onClick={() => {
                setTxHash('');
                setInputAmount('');
                clearEstimate();
              }}
              className="w-full"
            >
              Swap Again
            </Button>
          </div>
        </Card>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-gray-50">
      <header className="bg-white border-b border-gray-200">
        <div className="max-w-2xl mx-auto px-4 py-4">
          <button
            onClick={() => navigate('/dex')}
            className="flex items-center gap-2 text-gray-600 hover:text-gray-900"
          >
            <ArrowLeftIcon className="h-5 w-5" />
            Back to DEX
          </button>
        </div>
      </header>

      <main className="max-w-2xl mx-auto px-4 py-8">
        <Card>
          <h2 className="text-2xl font-bold mb-6">Swap Tokens</h2>

          {error && (
            <div className="bg-red-50 border border-red-200 rounded-lg p-3 mb-4 flex items-start gap-2">
              <ExclamationTriangleIcon className="h-5 w-5 text-red-600 flex-shrink-0 mt-0.5" />
              <p className="text-sm text-red-800">{error}</p>
            </div>
          )}

          <div className="space-y-4">
            {/* Input Asset */}
            <div>
              <AssetSelector
                label="From"
                selected={inputDenom}
                onSelect={setInputDenom}
                exclude={outputDenom || undefined}
              />
              <div className="mt-2 flex items-center gap-2">
                <input
                  type="number"
                  value={inputAmount}
                  onChange={(e) => setInputAmount(e.target.value)}
                  placeholder="0.00"
                  min="0"
                  step="any"
                  className="input flex-1"
                />
                <button
                  type="button"
                  onClick={handleSetMax}
                  className="text-sm text-primary-600 hover:text-primary-700 font-medium px-2"
                >
                  Max
                </button>
              </div>
              {inputBalance && (
                <div className="text-xs text-gray-500 mt-1">
                  Balance: {formatPnyx(inputBalance.amount)}{' '}
                  {inputAsset?.symbol || inputDenom}
                </div>
              )}
            </div>

            {/* Flip Button */}
            <div className="flex justify-center">
              <button
                type="button"
                onClick={handleFlipAssets}
                className="p-2 bg-gray-100 rounded-full hover:bg-gray-200 transition-colors"
              >
                <ArrowsUpDownIcon className="h-6 w-6 text-gray-600" />
              </button>
            </div>

            {/* Output Asset */}
            <div>
              <AssetSelector
                label="To"
                selected={outputDenom}
                onSelect={setOutputDenom}
                exclude={inputDenom || undefined}
              />
              {swapEstimate && (
                <div className="mt-2">
                  <div className="text-2xl font-bold">
                    {formatPnyx(swapEstimate.expected_output)}
                  </div>
                  <div className="text-xs text-gray-500">
                    {outputAsset?.symbol || outputDenom}
                  </div>
                </div>
              )}
            </div>

            {/* Slippage Tolerance */}
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                Slippage Tolerance
              </label>
              <div className="flex gap-2">
                {['0.5', '1', '2', '5'].map((value) => (
                  <button
                    key={value}
                    type="button"
                    onClick={() => setSlippage(value)}
                    className={`px-4 py-2 rounded-lg text-sm font-medium transition-colors ${
                      slippage === value
                        ? 'bg-primary-600 text-white'
                        : 'bg-gray-100 text-gray-700 hover:bg-gray-200'
                    }`}
                  >
                    {value}%
                  </button>
                ))}
                <input
                  type="number"
                  value={slippage}
                  onChange={(e) => setSlippage(e.target.value)}
                  className="input w-20"
                  placeholder="%"
                  min="0.1"
                  step="0.1"
                />
              </div>
            </div>
          </div>

          {/* Swap Details */}
          {swapEstimate && (
            <div className="mt-6 bg-gray-50 rounded-lg p-4 space-y-2">
              <div className="flex items-center justify-between text-sm">
                <span className="text-gray-600">Route</span>
                <span className="font-medium">
                  {swapEstimate.hops === 1 ? 'Direct' : `${swapEstimate.hops} hops`}
                </span>
              </div>
              {swapEstimate.route_symbols.length > 0 && (
                <div className="text-xs text-gray-500">
                  {swapEstimate.route_symbols.join(' → ')}
                </div>
              )}
              <div className="flex items-center justify-between text-sm">
                <span className="text-gray-600">Minimum Received</span>
                <span className="font-medium">
                  {formatPnyx(calculateMinOutput())}{' '}
                  {outputAsset?.symbol || outputDenom}
                </span>
              </div>
              <div className="flex items-center justify-between text-sm">
                <span className="text-gray-600">Slippage Tolerance</span>
                <span className="font-medium">{slippage}%</span>
              </div>
            </div>
          )}

          <Button
            onClick={handleExecuteSwap}
            isLoading={isSwapping}
            disabled={
              !inputDenom || !outputDenom || !inputAmount || !swapEstimate
            }
            className="w-full mt-6"
          >
            {swapEstimate ? 'Swap' : 'Enter Amount'}
          </Button>
        </Card>
      </main>
    </div>
  );
}
