#include <iostream>
#include <string>
#include <vector>

struct Result {
    int avgRating;
    int stones;
    std::string treasury;
    Result(int a, int s, std::string t) : avgRating(a), stones(s), treasury(t) {}
};

class TrueRepublicUI {
public:
    Result submitProposal(const std::string& domain, const std::string& issue, const std::string& suggestion, const std::string& voter) {
        std::cout << "Submitting: " << suggestion << " in " << domain << "/" << issue << " by " << voter << std::endl;
        return Result(0, 0, "500750");
    }

    Result rateProposal(const std::string& domain, const std::string& issue, const std::string& suggestion, const std::string& voter, int rating) {
        std::cout << voter << " rates " << suggestion << " with " << rating << std::endl;
        return Result(3, 5, "500700");
    }

    Result setStone(const std::string& domain, const std::string& issue, const std::string& suggestion, const std::string& voter) {
        std::cout << voter << " sets stone on " << suggestion << std::endl;
        return Result(3, 6, "500650");
    }
};

int main() {
    TrueRepublicUI ui;
    std::string domain = "PartyProgram", issue, suggestion, voter;
    int rating;

    std::cout << "Voter: "; std::cin >> voter;
    std::cout << "Issue: "; std::cin >> issue;
    std::cout << "Suggestion: "; std::getline(std::cin >> std::ws, suggestion);
    Result res = ui.submitProposal(domain, issue, suggestion, voter);
    std::cout << "Live: Avg Rating: " << res.avgRating << ", Stones: " << res.stones << ", Treasury: " << res.treasury << std::endl;

    std::cout << "Rate (-5 to +5): "; std::cin >> rating;
    res = ui.rateProposal(domain, issue, suggestion, voter, rating);
    std::cout << "Live: Avg Rating: " << res.avgRating << ", Stones: " << res.stones << ", Treasury: " << res.treasury << std::endl;

    res = ui.setStone(domain, issue, suggestion, voter);
    std::cout << "Live: Avg Rating: " << res.avgRating << ", Stones: " << res.stones << ", Treasury: " << res.treasury << std::endl;

    return 0;
}
