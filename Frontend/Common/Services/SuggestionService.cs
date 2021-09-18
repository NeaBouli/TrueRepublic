using System;
using System.Collections.Generic;
using System.Linq;
using Common.Data;
using Common.Entities;
using Microsoft.EntityFrameworkCore;

namespace Common.Services
{
    /// <summary>
    /// Implementation of the suggestion service
    /// </summary>
    public class SuggestionService
    {
        public List<Suggestion> GetSuggestionsForIssue(Guid issueId)
        {
            DbServiceContext dbServiceContext = DatabaseInitializationService.GetDbServiceContext();

            using (dbServiceContext)
            {
                Issue issue = dbServiceContext.Issues
                    .Include(i => i.Suggestions)
                    .FirstOrDefault(i => i.Id.ToString() == issueId.ToString());

                if (issue == null)
                {
                    return null;
                }

                SuggestionService.UpdateStakes(issue.Suggestions);

                return issue.Suggestions;
            }
        }

        /// <summary>
        /// Rolls the back invalid staked suggestions.
        /// </summary>
        public void RollBackInvalidStakedSuggestions()
        {
            DbServiceContext dbServiceContext = DatabaseInitializationService.GetDbServiceContext();

            using (dbServiceContext)
            {
                List<StakedSuggestion> invalidStakedSuggestions =
                    dbServiceContext.StakedSuggestions.Where(s => s.IsExpired).ToList();

                if (invalidStakedSuggestions.Count == 0)
                {
                    return;
                }

                List<Wallet> wallets = dbServiceContext.Wallets.ToList();

                foreach (StakedSuggestion stakedSuggestion in invalidStakedSuggestions)
                {
                    foreach (Wallet wallet in wallets)
                    {
                        List<WalletTransaction> walletTransactionsToAdd = new List<WalletTransaction>();
                        double balanceChange = 0;

                        foreach (WalletTransaction walletTransaction in wallet.WalletTransactions)
                        {
                            if (walletTransaction.TransactionId != null &&
                                walletTransaction.TransactionId.ToString() == stakedSuggestion.Id.ToString() &&
                                walletTransaction.TransactionType.Name ==
                                TransactionTypeNames.StakeSuggestion.ToString())
                            {
                                WalletTransaction walletTransactionToAdd = new WalletTransaction
                                {
                                    Balance = (-1) * walletTransaction.Balance,
                                    CreateDate = DateTime.Now,
                                    TransactionId = walletTransaction.TransactionId
                                };
                                walletTransactionToAdd.TransactionType = new TransactionType
                                {
                                    Name = TransactionTypeNames.StakeSuggestionRollback.ToString(),
                                    Id = Guid.NewGuid(),
                                    Fee = walletTransactionToAdd.Balance
                                };

                                balanceChange += walletTransactionToAdd.Balance;
                                walletTransactionsToAdd.Add(walletTransactionToAdd);
                            }
                        }

                        wallet.WalletTransactions.AddRange(walletTransactionsToAdd);
                        wallet.TotalBalance += balanceChange;
                    }
                }
            }
        }

        /// <summary>
        /// Updates the stakes.
        /// </summary>
        /// <param name="suggestions">The suggestions.</param>
        public static void UpdateStakes(List<Suggestion> suggestions)
        {
            DbServiceContext dbServiceContext = DatabaseInitializationService.GetDbServiceContext();

            using (dbServiceContext)
            {
                foreach (Suggestion suggestion in suggestions)
                {
                    int count = dbServiceContext.StakedSuggestions
                        .Include(s => s.Suggestion)
                        .Where(s => s.Suggestion.Id.ToString() == suggestion.Id.ToString())
                        .ToList().Count;

                    suggestion.StakeCount = count;
                }
            }
        }
    }
}
