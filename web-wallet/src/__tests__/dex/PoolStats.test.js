import React from "react";
import { render, screen, waitFor } from "@testing-library/react";
import PoolStats from "../../components/dex/PoolStats";
import { queryPoolStats } from "../../services/api";

jest.mock("../../services/api", () => ({
  queryPoolStats: jest.fn(),
}));

describe("PoolStats", () => {
  beforeEach(() => jest.clearAllMocks());

  test("shows loading state", () => {
    queryPoolStats.mockReturnValue(new Promise(() => {}));
    render(<PoolStats assetDenom="atom" />);
    expect(screen.getByText(/loading/i)).toBeInTheDocument();
  });

  test("renders pool statistics", async () => {
    queryPoolStats.mockResolvedValue({
      asset_denom: "atom",
      asset_symbol: "ATOM",
      swap_count: 42,
      total_volume_pnyx: "1000000",
      total_fees_earned: "30",
      total_burned: "100",
      pnyx_reserve: "500000",
      asset_reserve: "500000",
      spot_price_per_million: "997000",
      total_shares: "1000000",
    });
    render(<PoolStats assetDenom="atom" />);
    await waitFor(() => {
      expect(screen.getByText("42")).toBeInTheDocument();
    });
  });

  test("shows error message on failure", async () => {
    queryPoolStats.mockRejectedValue(new Error("Pool not found"));
    render(<PoolStats assetDenom="invalid" />);
    await waitFor(() => {
      expect(screen.getByText(/pool not found/i)).toBeInTheDocument();
    });
  });
});
