import React from "react";
import { render, screen, fireEvent, waitFor } from "@testing-library/react";
import ZKPVotingPanel from "../../components/zkp/ZKPVotingPanel";
import { submitAnonymousVote } from "../../services/api";

jest.mock("../../services/api", () => ({
  submitAnonymousVote: jest.fn(),
}));

// Mock crypto.getRandomValues for Node test environment.
const mockCrypto = {
  getRandomValues: (arr) => {
    for (let i = 0; i < arr.length; i++) arr[i] = i % 256;
    return arr;
  },
};
Object.defineProperty(global, "crypto", { value: mockCrypto });

describe("ZKPVotingPanel", () => {
  const defaultProps = {
    domainName: "TestDomain",
    issueName: "Climate",
    suggestionName: "GreenDeal",
    connected: true,
    address: "truerepublic1abc",
  };

  beforeEach(() => jest.clearAllMocks());

  test("renders slider with default value 0", () => {
    render(<ZKPVotingPanel {...defaultProps} />);
    expect(screen.getByText("0")).toBeInTheDocument();
    expect(screen.getByRole("slider")).toHaveValue("0");
  });

  test("renders generate proof button", () => {
    render(<ZKPVotingPanel {...defaultProps} />);
    expect(screen.getByText("Generate ZKP Proof")).toBeInTheDocument();
  });

  test("proof generation shows submit button", async () => {
    render(<ZKPVotingPanel {...defaultProps} />);
    fireEvent.click(screen.getByText("Generate ZKP Proof"));
    await waitFor(
      () => {
        expect(screen.getByText("Submit Anonymous Vote")).toBeInTheDocument();
      },
      { timeout: 3000 }
    );
  });

  test("shows disconnect message when not connected", () => {
    render(<ZKPVotingPanel {...defaultProps} connected={false} />);
    expect(
      screen.getByText("Connect wallet for anonymous voting")
    ).toBeInTheDocument();
  });
});
