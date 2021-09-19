using System;
using System.Collections.Generic;
using System.Data;
using System.Linq;
using Common.Data;
using Common.Entities;
using Microsoft.EntityFrameworkCore;

namespace Common.Services
{
    public class StakedSuggestionService
    {
        /// <summary>
        /// Gets the staked suggestions for user.
        /// </summary>
        /// <param name="userId">The user identifier.</param>
        /// <returns>The staked suggestions for the user</returns>
        public List<StakedSuggestion> GetStakedSuggestionsForUser(Guid userId)
        {
            RollBackInvalidStakedSuggestions();

            DbServiceContext dbServiceContext = DatabaseInitializationService.GetDbServiceContext();

            using (dbServiceContext)
            {
                return dbServiceContext.User
                    .FirstOrDefault(u => u.Id.ToString() == userId.ToString())?
                    .StakedSuggestions;
            }
        }

        /// <summary>
        /// Stakes the specified suggestion identifier.
        /// </summary>
        /// <param name="suggestionId">The suggestion identifier.</param>
        /// <param name="userId">The user identifier.</param>
        /// <exception cref="System.InvalidOperationException"></exception>
        public void Stake(Guid suggestionId, Guid userId)
        {
            List<StakedSuggestion> stakedSuggestionsForUser = GetStakedSuggestionsForUser(userId);

            if (stakedSuggestionsForUser.Any(stakedSuggestion => stakedSuggestion.Suggestion.Id.ToString() == suggestionId.ToString()))
            {
                throw new InvalidOperationException(Resource.ErrorSuggestionAlreadyStakedForUser);
            }

            WalletService walletService = new WalletService();
            walletService.AddTransaction(userId, TransactionTypeNames.StakeSuggestion, suggestionId);
        }

        /// <summary>
        /// Rolls the back invalid staked suggestions.
        /// </summary>
        public void RollBackInvalidStakedSuggestions()
        {
            // TODO: move into StakedSuggestionService
            // TODO: under which conditions is this triggered and why is this needed?

            // TODO: would be much easier if we track the user id

            DbServiceContext dbServiceContext = DatabaseInitializationService.GetDbServiceContext();

            using (dbServiceContext)
            {
                List<StakedSuggestion> invalidStakedSuggestions =
                    dbServiceContext.StakedSuggestions.Where(s => s.IsExpired).ToList();

                if (invalidStakedSuggestions.Count == 0)
                {
                    return;
                }

                WalletService walletService = new WalletService();

                foreach (StakedSuggestion stakedSuggestion in invalidStakedSuggestions)
                {
                    // TODO: implementation
                }

                throw new NotImplementedException();

                /*
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
                */
            }
        }

        public int Import(DataTable dataTable)
        {
            throw new NotImplementedException();
        }
    }
}
