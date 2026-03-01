import React from "react";
import { render, screen, waitFor } from "@testing-library/react";
import MembershipStatus from "../../components/zkp/MembershipStatus";
import { queryZKPState } from "../../services/api";

jest.mock("../../services/api", () => ({
  queryZKPState: jest.fn(),
}));

describe("MembershipStatus", () => {
  beforeEach(() => jest.clearAllMocks());

  test("shows loading state initially", () => {
    queryZKPState.mockReturnValue(new Promise(() => {}));
    render(<MembershipStatus domainName="TestDomain" />);
    expect(screen.getByText(/loading/i)).toBeInTheDocument();
  });

  test("displays membership data", async () => {
    queryZKPState.mockResolvedValue({
      merkle_root: "0xabcdef1234567890abcdef1234567890",
      commitment_count: 5,
      member_count: 10,
      vk_initialized: true,
    });
    render(<MembershipStatus domainName="TestDomain" />);
    await waitFor(() => {
      expect(screen.getByText("10")).toBeInTheDocument();
      expect(screen.getByText("5")).toBeInTheDocument();
    });
  });

  test("shows error on failure", async () => {
    queryZKPState.mockRejectedValue(new Error("Network error"));
    render(<MembershipStatus domainName="TestDomain" />);
    await waitFor(() => {
      expect(screen.getByText(/network error/i)).toBeInTheDocument();
    });
  });
});
