import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { useToastStore } from './toastStore';

describe('toast store', () => {
  beforeEach(() => {
    vi.useFakeTimers();
    vi.spyOn(crypto, 'randomUUID').mockReturnValue(
      '00000000-0000-4000-8000-000000000000'
    );
    useToastStore.setState({ toasts: [] });
  });

  afterEach(() => {
    vi.restoreAllMocks();
    vi.useRealTimers();
  });

  it('adds and explicitly removes a toast', () => {
    useToastStore.getState().addToast({
      type: 'success',
      message: 'Saved',
      duration: 0,
    });

    expect(useToastStore.getState().toasts).toEqual([
      {
        id: '00000000-0000-4000-8000-000000000000',
        type: 'success',
        message: 'Saved',
        duration: 0,
      },
    ]);

    useToastStore
      .getState()
      .removeToast('00000000-0000-4000-8000-000000000000');
    expect(useToastStore.getState().toasts).toEqual([]);
  });

  it('removes a toast after the default duration', () => {
    useToastStore.getState().addToast({ type: 'info', message: 'Working' });

    vi.advanceTimersByTime(4_999);
    expect(useToastStore.getState().toasts).toHaveLength(1);

    vi.advanceTimersByTime(1);
    expect(useToastStore.getState().toasts).toHaveLength(0);
  });
});
