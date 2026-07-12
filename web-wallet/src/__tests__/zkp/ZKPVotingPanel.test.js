import React from "react";
import { render, screen } from "@testing-library/react";
import ZKPVotingPanel from "../../components/zkp/ZKPVotingPanel";
import { submitAnonymousVote } from "../../services/api";

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

  test("clearly disables the mock prover", () => {
    render(<ZKPVotingPanel {...defaultProps} />);
    const button = screen.getByText("Real ZKP Prover Unavailable");
    expect(button).toBeDisabled();
    expect(screen.queryByText("Submit Anonymous Vote")).not.toBeInTheDocument();
    expect(screen.getByText(/Submission disabled/)).toBeInTheDocument();
  });

  test("shows disconnect message when not connected", () => {
    render(<ZKPVotingPanel {...defaultProps} connected={false} />);
    expect(
      screen.getByText("Connect wallet for anonymous voting")
    ).toBeInTheDocument();
  });

  test("API rejects mock proof submission fail-closed", async () => {
    await expect(
      submitAnonymousVote(
        defaultProps.address,
        defaultProps.domainName,
        defaultProps.issueName,
        defaultProps.suggestionName,
        3,
        "00",
        "00",
        ""
      )
    ).rejects.toThrow("submission is disabled");
  });
});
