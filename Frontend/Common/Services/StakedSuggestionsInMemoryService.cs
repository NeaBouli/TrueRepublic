using System;
using System.Collections.Generic;
using System.Linq;
using Common.Entities;

namespace Common.Services
{
    /// <summary>
    /// Implementation of the StakedSuggestionsInMemoryService
    /// </summary>
    public class StakedSuggestionsInMemoryService
    {
        /// <summary>
        /// The staked suggestions
        /// </summary>
        private readonly List<StakedSuggestion> _stakedSuggestions;

        /// <summary>
        /// The wallets
        /// </summary>
        private readonly List<Wallet> _wallets;

        /// <summary>
        /// Initializes a new instance of the <see cref="StakedSuggestionsInMemoryService"/> class.
        /// </summary>
        /// <param name="stakedSuggestions">The staked suggestions.</param>
        /// <param name="wallets">The wallets.</param>
        public StakedSuggestionsInMemoryService(List<StakedSuggestion> stakedSuggestions, List<Wallet> wallets)
        {
            _stakedSuggestions = stakedSuggestions;
            _wallets = wallets;
        }

        public List<StakedSuggestion> GetStakedSuggestions()
        {
            return _stakedSuggestions;
        }

        public List<StakedSuggestion> GetValidStakedSuggestions()
        {
            return _stakedSuggestions.Where(s => s.ValidTill >= DateTime.Now).ToList();
        }

        public List<StakedSuggestion> GetInvalidStakedSuggestions()
        {
            return _stakedSuggestions.Where(s => s.ValidTill < DateTime.Now).ToList();
        }

        /// <summary>
        /// Rolls the back invalid staked suggestions.
        /// </summary>
        public void RollBackInvalidStakedSuggestions()
        {
            List<StakedSuggestion> invalidStakedSuggestions = GetInvalidStakedSuggestions();

            if (invalidStakedSuggestions.Count == 0)
            {
                return;
            }

            foreach (StakedSuggestion stakedSuggestion in invalidStakedSuggestions)
            {
                foreach (Wallet wallet in _wallets)
                {
                    List<WalletTransaction> walletTransactionsToAdd = new List<WalletTransaction>();
                    double balanceChange = 0;

                    foreach (WalletTransaction walletTransaction in wallet.WalletTransactions)
                    {
                        if (walletTransaction.TransactionId != null &&
                            walletTransaction.TransactionId.ToString() == stakedSuggestion.Id.ToString() &&
                            walletTransaction.TransactionType.Name == TransactionTypeNames.StakeSuggestion.ToString())
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
}
