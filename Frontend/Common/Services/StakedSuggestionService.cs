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
        /// <exception cref="System.InvalidOperationException">ErrorSuggestionAlreadyStakedForUser</exception>
        public void Stake(Guid suggestionId, Guid userId)
        {
            // TODO: refactor

            // old code
            List<StakedSuggestion> stakedSuggestionsForUser = GetStakedSuggestionsForUser(userId);

            if (stakedSuggestionsForUser.Any(stakedSuggestion =>
                stakedSuggestion.Suggestion.Id.ToString() == suggestionId.ToString()))
            {
                throw new InvalidOperationException(Resource.ErrorSuggestionAlreadyStakedForUser);
            }

            WalletService walletService = new WalletService();

            DbServiceContext dbServiceContext = DatabaseInitializationService.GetDbServiceContext();

            using (dbServiceContext)
            {

                Guid? issueId =
                    stakedSuggestionsForUser.FirstOrDefault(s => s.Suggestion.Id.ToString() == suggestionId.ToString())
                        ?.IssueId;

                if (issueId != null)
                {
                    StakedSuggestion existingStakedSuggestion =
                        stakedSuggestionsForUser.FirstOrDefault(
                            s => s.IssueId.ToString() == issueId.ToString() &&
                                 suggestionId.ToString() != s.Suggestion.Id.ToString());

                    if (existingStakedSuggestion != null)
                    {
                        walletService.AddTransaction(userId, TransactionTypeNames.StakeSuggestionRollback,
                            suggestionId);


                        dbServiceContext.Remove(existingStakedSuggestion);
                    }
                }

                walletService.AddTransaction(userId, TransactionTypeNames.StakeSuggestion, suggestionId);

                // TODO: add staked suggestion
                StakedSuggestion stakedSuggestion = new StakedSuggestion();
                stakedSuggestion.Suggestion =
                    dbServiceContext.Suggestions.FirstOrDefault(s => s.Id.ToString() == suggestionId.ToString());
                if (stakedSuggestion.Suggestion != null)
                {
                    stakedSuggestion.IssueId = stakedSuggestion.Suggestion.IssueId;
                }

                User user = dbServiceContext.User.FirstOrDefault(u => u.Id.ToString() == userId.ToString());
                if (user != null)
                {
                    user.StakedSuggestions.Add(stakedSuggestion);
                }

                dbServiceContext.SaveChanges();
            }
        }

        /// <summary>
        /// Rolls the back invalid staked suggestions.
        /// </summary>
        public void RollBackInvalidStakedSuggestions()
        {
            // TODO: under which conditions is this triggered and why is this needed?
            // TODO: refactor

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
