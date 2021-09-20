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

                foreach (User user in dbServiceContext.User.Include(u => u.StakedSuggestions).ToList())
                {
                    foreach (StakedSuggestion stakedSuggestion in invalidStakedSuggestions)
                    {
                        if (!user.StakedSuggestions.Contains(stakedSuggestion))
                        {
                            continue;
                        }
                        
                        Guid userId = user.Id;
                        Guid suggestionId = stakedSuggestion.Suggestion.Id;
                        walletService.AddTransaction(userId, TransactionTypeNames.StakeSuggestionRollback, suggestionId);
                    }
                }
            }
        }

        public int Import(DataTable dataTable)
        {
            throw new NotImplementedException();
        }
    }
}
