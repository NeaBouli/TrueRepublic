import React from "react";
import { render, screen, fireEvent, waitFor } from "@testing-library/react";
import NullifierStatus from "../../components/zkp/NullifierStatus";
import { queryNullifier } from "../../services/api";

jest.mock("../../services/api", () => ({
  queryNullifier: jest.fn(),
}));

describe("NullifierStatus", () => {
  beforeEach(() => jest.clearAllMocks());

  test("renders input and check button", () => {
    render(<NullifierStatus domainName="TestDomain" />);
    expect(screen.getByPlaceholderText(/nullifier hash/i)).toBeInTheDocument();
    expect(screen.getByText("Check")).toBeInTheDocument();
  });

  test("shows available for unused nullifier", async () => {
    queryNullifier.mockResolvedValue({ used: false });
    render(<NullifierStatus domainName="TestDomain" />);
    fireEvent.change(screen.getByPlaceholderText(/nullifier hash/i), {
      target: { value: "abc123" },
    });
    fireEvent.click(screen.getByText("Check"));
    await waitFor(() => {
      expect(screen.getByText(/available/i)).toBeInTheDocument();
    });
  });

  test("shows used for spent nullifier", async () => {
    queryNullifier.mockResolvedValue({ used: true });
    render(<NullifierStatus domainName="TestDomain" />);
    fireEvent.change(screen.getByPlaceholderText(/nullifier hash/i), {
      target: { value: "used123" },
    });
    fireEvent.click(screen.getByText("Check"));
    await waitFor(() => {
      expect(screen.getByText(/already used/i)).toBeInTheDocument();
    });
  });
});
