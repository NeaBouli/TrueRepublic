import { useEffect, useMemo, useState } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import { ArrowLeftIcon, ArrowPathIcon, PaperAirplaneIcon } from '@heroicons/react/24/outline';
import { Button } from '@/components/common/Button';
import { DEFAULT_CHAIN } from '@/config/chains';
import {
  ibcTransferScopeKey,
  useIbcTransferStore,
} from '@/stores/ibcTransferStore';
import { useWalletStore } from '@/stores/walletStore';
import type { IbcTransferPhase } from '@/types/ibc';
import { formatPnyx } from '@/utils/format';

const phaseCopy: Record<IbcTransferPhase, { label: string; tone: string; detail: string }> = {
  validating: { label: 'Validating', tone: 'bg-slate-100 text-slate-700', detail: 'No signature has been requested.' },
  signing: { label: 'Signing', tone: 'bg-blue-100 text-blue-800', detail: 'Confirm the exact transfer in your wallet.' },
  broadcasting: { label: 'Broadcasting', tone: 'bg-blue-100 text-blue-800', detail: 'Submission is in progress; do not send again.' },
  committed_pending_relay: { label: 'Committed · pending relay', tone: 'bg-amber-100 text-amber-900', detail: 'The source chain committed the packet. Delivery is not yet proven.' },
  acknowledged: { label: 'Acknowledged', tone: 'bg-emerald-100 text-emerald-900', detail: 'Source-chain evidence proves a successful acknowledgement.' },
  acknowledged_error_or_refunded: { label: 'Error acknowledgement · refund path', tone: 'bg-rose-100 text-rose-900', detail: 'The source chain recorded an error acknowledgement and the transfer module refund path.' },
  timed_out_refunded: { label: 'Timed out · refund path', tone: 'bg-rose-100 text-rose-900', detail: 'The source chain recorded a timeout and the transfer module refund path.' },
  unknown: { label: 'Unknown · check before retrying', tone: 'bg-slate-200 text-slate-900', detail: 'Evidence is missing or contradictory. Never resubmit until the transaction is checked.' },
};

export function IbcTransferPage() {
  const navigate = useNavigate();
  const { txHash } = useParams();
  const { currentWallet, isLocked, balances } = useWalletStore();
  const store = useIbcTransferStore();
  const [channelId, setChannelId] = useState('');
  const [receiver, setReceiver] = useState('');
  const [amount, setAmount] = useState('');
  const [memo, setMemo] = useState('');
  const [confirmedReceiver, setConfirmedReceiver] = useState(false);
  const [formError, setFormError] = useState<string | null>(null);

  const address = currentWallet?.address ?? '';
  const records = useMemo(
    () =>
      address
        ? store.recordsByScope[ibcTransferScopeKey(DEFAULT_CHAIN.chainId, address)] ?? []
        : [],
    [address, store.recordsByScope]
  );
  const selectedRecord = txHash
    ? records.find((record) => record.txHash === txHash.toUpperCase())
    : undefined;

  useEffect(() => {
    if (isLocked || !currentWallet) {
      navigate('/unlock', { replace: true });
      return;
    }
    void store.loadChannels().catch(() => undefined);
    // Store actions are stable; wallet changes intentionally re-run discovery.
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [address, isLocked, navigate]);

  if (!currentWallet) return null;

  const nativeBalance = balances.find((balance) => balance.denom === 'upnyx')?.amount ?? '0';
  const effectiveChannelId = channelId || store.channels[0]?.channelId || '';
  const phase = selectedRecord?.phase ?? store.submissionPhase;

  const submit = async (event: React.FormEvent) => {
    event.preventDefault();
    setFormError(null);
    const channel = store.channels.find((candidate) => candidate.channelId === effectiveChannelId);
    if (!channel) return setFormError('Select a verified open ICS-20 channel.');
    if (!confirmedReceiver) {
      return setFormError('Confirm that you verified the receiver on the counterparty chain.');
    }
    try {
      const record = await store.submitTransfer({ channel, receiver, amount, memo });
      navigate(`/ibc/transfer/${record.txHash}`);
    } catch (error) {
      setFormError(error instanceof Error ? error.message : 'IBC transfer failed');
    }
  };

  const reconcile = async (hash: string) => {
    setFormError(null);
    try {
      await store.reconcileTransfer(hash);
    } catch (error) {
      setFormError(error instanceof Error ? error.message : 'Recovery check failed');
    }
  };

  return (
    <main className="min-h-screen bg-slate-950 text-slate-100">
      <div className="mx-auto max-w-5xl px-4 py-8">
        <button onClick={() => navigate('/wallet')} className="mb-6 inline-flex items-center gap-2 text-sm text-slate-300 hover:text-white">
          <ArrowLeftIcon className="h-4 w-4" /> Back to wallet
        </button>

        <header className="mb-8 border-l-4 border-cyan-400 pl-5">
          <p className="font-mono text-xs uppercase tracking-[0.22em] text-cyan-300">Cross-chain operations</p>
          <h1 className="mt-2 text-3xl font-bold">IBC transfer & recovery</h1>
          <p className="mt-2 max-w-2xl text-sm text-slate-300">
            Broadcast is not delivery. This screen reports only source-chain transaction,
            acknowledgement, and timeout evidence.
          </p>
        </header>

        <div className="grid gap-6 lg:grid-cols-[1.1fr_0.9fr]">
          <form onSubmit={submit} className="rounded-2xl border border-slate-700 bg-slate-900 p-6 shadow-xl">
            <div className="mb-6 flex items-start justify-between gap-4">
              <div>
                <h2 className="text-lg font-semibold">New native transfer</h2>
                <p className="mt-1 text-xs text-slate-400">Balance: {formatPnyx(nativeBalance)} PNYX · 0.01 PNYX held as fee reserve</p>
              </div>
              <span className="rounded-full bg-cyan-950 px-3 py-1 font-mono text-xs text-cyan-200">ICS-20</span>
            </div>

            {store.channelStatus === 'loading' && <p role="status" className="mb-4 text-sm text-slate-300">Checking open transfer channels…</p>}
            {store.channelStatus === 'error' && (
              <div role="alert" className="mb-4 rounded-lg border border-rose-700 bg-rose-950/50 p-3 text-sm">
                Channel query failed ({store.channelFailure ?? 'unknown'}). This is not an empty channel set.
                <button type="button" onClick={() => void store.loadChannels().catch(() => undefined)} className="ml-2 underline">Retry query</button>
              </div>
            )}
            {store.channelStatus === 'ready' && store.channels.length === 0 && (
              <p role="status" className="mb-4 rounded-lg border border-amber-700 bg-amber-950/40 p-3 text-sm">The chain returned no selectable open ICS-20 transfer channel.</p>
            )}

            <label className="mb-4 block text-sm font-medium">
              Source channel
              <select value={effectiveChannelId} onChange={(event) => setChannelId(event.target.value)} disabled={store.channels.length === 0} className="mt-2 w-full rounded-lg border border-slate-600 bg-slate-950 px-3 py-3">
                {store.channels.length === 0 && <option value="">No channel available</option>}
                {store.channels.map((channel) => <option key={channel.channelId} value={channel.channelId}>{channel.channelId} → {channel.counterpartyChannelId}</option>)}
              </select>
            </label>
            <label className="mb-4 block text-sm font-medium">
              Counterparty receiver
              <input value={receiver} onChange={(event) => setReceiver(event.target.value)} required autoComplete="off" className="mt-2 w-full rounded-lg border border-slate-600 bg-slate-950 px-3 py-3 font-mono text-sm" placeholder="cosmos1…" />
            </label>
            <label className="mb-4 block text-sm font-medium">
              Amount (PNYX)
              <input value={amount} onChange={(event) => setAmount(event.target.value)} required inputMode="decimal" className="mt-2 w-full rounded-lg border border-slate-600 bg-slate-950 px-3 py-3" placeholder="1.000000" />
            </label>
            <label className="mb-5 block text-sm font-medium">
              Memo <span className="text-slate-500">(optional)</span>
              <input value={memo} onChange={(event) => setMemo(event.target.value)} maxLength={256} className="mt-2 w-full rounded-lg border border-slate-600 bg-slate-950 px-3 py-3" />
            </label>
            <label className="mb-5 flex items-start gap-3 rounded-lg bg-slate-800 p-3 text-sm">
              <input type="checkbox" checked={confirmedReceiver} onChange={(event) => setConfirmedReceiver(event.target.checked)} className="mt-1 h-4 w-4" />
              <span>I verified this receiver and address prefix on the intended counterparty chain. Bech32 syntax alone does not prove the destination network.</span>
            </label>
            {(formError || store.submissionError || store.reconcileError) && <p role="alert" className="mb-4 text-sm text-rose-300">{formError ?? store.submissionError ?? store.reconcileError}</p>}
            <Button type="submit" isLoading={store.submissionStatus === 'loading'} disabled={store.channelStatus !== 'ready' || store.channels.length === 0} className="w-full justify-center">
              <PaperAirplaneIcon className="h-5 w-5" /> Validate & sign once
            </Button>
          </form>

          <section aria-labelledby="transfer-ledger-title" className="rounded-2xl border border-slate-700 bg-slate-900 p-6">
            <h2 id="transfer-ledger-title" className="text-lg font-semibold">Transfer evidence ledger</h2>
            {phase && <div className="mt-4 rounded-xl border border-slate-700 p-4"><span className={`inline-flex rounded-full px-3 py-1 text-xs font-semibold ${phaseCopy[phase].tone}`}>{phaseCopy[phase].label}</span><p className="mt-3 text-sm text-slate-300">{phaseCopy[phase].detail}</p></div>}
            <div className="mt-5 space-y-3">
              {records.length === 0 && <p className="text-sm text-slate-400">No transfer records for this wallet and chain.</p>}
              {records.map((record) => (
                <article key={record.txHash} className={`rounded-xl border p-4 ${selectedRecord?.txHash === record.txHash ? 'border-cyan-500 bg-cyan-950/20' : 'border-slate-700'}`}>
                  <div className="flex items-start justify-between gap-3">
                    <button onClick={() => navigate(`/ibc/transfer/${record.txHash}`)} className="min-w-0 text-left">
                      <span className="block truncate font-mono text-xs text-cyan-300">{record.txHash}</span>
                      <span className="mt-1 block text-sm">{formatPnyx(record.amount)} PNYX · {record.channel.channelId}</span>
                    </button>
                    <span className={`shrink-0 rounded-full px-2 py-1 text-[11px] font-semibold ${phaseCopy[record.phase].tone}`}>{phaseCopy[record.phase].label}</span>
                  </div>
                  <button type="button" onClick={() => void reconcile(record.txHash)} disabled={!record.packet || store.reconcileStatus === 'loading'} className="mt-3 inline-flex items-center gap-2 text-xs text-slate-300 underline disabled:cursor-not-allowed disabled:text-slate-600">
                    <ArrowPathIcon className="h-4 w-4" /> Check source-chain evidence
                  </button>
                  {!record.packet && <p className="mt-2 text-xs text-amber-300">No send_packet evidence is stored. Check the transaction hash before any retry.</p>}
                </article>
              ))}
            </div>
          </section>
        </div>
      </div>
    </main>
  );
}
