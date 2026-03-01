import React from "react";
import { render, screen, fireEvent, waitFor } from "@testing-library/react";
import SwapEstimate from "../../components/dex/SwapEstimate";
import { queryEstimateSwap } from "../../services/api";

jest.mock("../../services/api", () => ({
  queryEstimateSwap: jest.fn(),
}));

describe("SwapEstimate", () => {
  const pools = [{ asset_denom: "atom" }];

  beforeEach(() => jest.clearAllMocks());

  test("renders estimate form", () => {
    render(<SwapEstimate pools={pools} />);
    expect(screen.getByText("Swap Estimate")).toBeInTheDocument();
    expect(screen.getByText("Estimate Output")).toBeInTheDocument();
  });

  test("shows estimate result with route", async () => {
    queryEstimateSwap.mockResolvedValue({
      expected_output: "990",
      route: ["pnyx", "atom"],
      route_symbols: ["PNYX", "ATOM"],
      hops: 1,
    });
    render(<SwapEstimate pools={pools} />);

    fireEvent.change(screen.getByPlaceholderText("0"), {
      target: { value: "1000" },
    });
    fireEvent.click(screen.getByText("Estimate Output"));

    await waitFor(() => {
      expect(screen.getByText(/990/)).toBeInTheDocument();
    });
  });

  test("shows cross-asset warning for multi-hop", async () => {
    queryEstimateSwap.mockResolvedValue({
      expected_output: "950",
      route: ["atom", "pnyx", "btc"],
      route_symbols: ["ATOM", "PNYX", "BTC"],
      hops: 2,
    });
    render(<SwapEstimate pools={[{ asset_denom: "atom" }, { asset_denom: "btc" }]} />);

    fireEvent.change(screen.getByPlaceholderText("0"), {
      target: { value: "1000" },
    });
    fireEvent.click(screen.getByText("Estimate Output"));

    await waitFor(() => {
      expect(screen.getByText(/cross-asset swap/i)).toBeInTheDocument();
    });
  });
});
