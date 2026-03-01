import React from "react";
import { render, screen, fireEvent, waitFor } from "@testing-library/react";
import SpotPriceDisplay from "../../components/dex/SpotPriceDisplay";
import { querySpotPrice } from "../../services/api";

jest.mock("../../services/api", () => ({
  querySpotPrice: jest.fn(),
}));

describe("SpotPriceDisplay", () => {
  const pools = [{ asset_denom: "atom" }, { asset_denom: "btc" }];

  beforeEach(() => jest.clearAllMocks());

  test("renders denom selectors", () => {
    render(<SpotPriceDisplay pools={pools} />);
    expect(screen.getByText("Spot Price")).toBeInTheDocument();
    expect(screen.getByText("Get Price")).toBeInTheDocument();
  });

  test("fetches and displays price", async () => {
    querySpotPrice.mockResolvedValue({
      input_symbol: "PNYX",
      output_symbol: "ATOM",
      price_per_million: "997000",
      route: ["pnyx", "atom"],
    });
    render(<SpotPriceDisplay pools={pools} />);
    fireEvent.click(screen.getByText("Get Price"));
    await waitFor(() => {
      expect(screen.getByText(/0.997000/)).toBeInTheDocument();
    });
  });
});
