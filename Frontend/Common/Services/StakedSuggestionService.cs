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
        /// The expiration days
        /// </summary>
        private readonly int _expirationDays;

        /// <summary>
        /// Initializes a new instance of the <see cref="StakedSuggestionService"/> class.
        /// </summary>
        public StakedSuggestionService()
        {
            _expirationDays = 0;
        }

        /// <summary>
        /// Initializes a new instance of the <see cref="StakedSuggestionService"/> class.
        /// </summary>
        /// <param name="expirationDays">The expiration days.</param>
        public StakedSuggestionService(int expirationDays)
        {
            _expirationDays = expirationDays;
        }

        public List<StakedSuggestion> GetStakedSuggestionsForUser(DbServiceContext dbServiceContext, Guid userId)
        {
            RollBackInvalidStakedSuggestions(dbServiceContext);

            return dbServiceContext.User
                .FirstOrDefault(u => u.Id.ToString() == userId.ToString())?
                .StakedSuggestions;
        }

        /// <summary>
        /// Stakes the specified suggestion identifier.
        /// </summary>
        /// <param name="dbServiceContext">The database service context.</param>
        /// <param name="suggestionId">The suggestion identifier.</param>
        /// <param name="userId">The user identifier.</param>
        /// <exception cref="System.InvalidOperationException">ErrorSuggestionAlreadyStakedForUser</exception>
        public void Stake(DbServiceContext dbServiceContext, Guid suggestionId, Guid userId)
        {
            if (_expirationDays == 0)
            {
                throw new InvalidOperationException(Resource.ErrorExpirationDaysNeedsToBeSet);
            }

            List<StakedSuggestion> stakedSuggestionsForUser = GetStakedSuggestionsForUser(dbServiceContext, userId);

            ThrowExceptionIfAlreadyStaked(suggestionId, stakedSuggestionsForUser);

            WalletService walletService = new WalletService();

            RollbackOtherStakeForIssueIfAlreadyExisting(dbServiceContext, walletService, suggestionId, userId,
                stakedSuggestionsForUser);

            walletService.AddTransaction(dbServiceContext, userId, TransactionTypeNames.StakeSuggestion, suggestionId);

            AddStakeSuggestion(dbServiceContext, suggestionId, userId);

            dbServiceContext.SaveChanges();
        }

        /// <summary>
        /// Adds the stake suggestion.
        /// </summary>
        /// <param name="dbServiceContext">The database service context.</param>
        /// <param name="suggestionId">The suggestion identifier.</param>
        /// <param name="userId">The user identifier.</param>
        /// <exception cref="System.InvalidOperationException"></exception>
        private void AddStakeSuggestion(DbServiceContext dbServiceContext, Guid suggestionId, Guid userId)
        {
            if (_expirationDays == 0)
            {
                throw new InvalidOperationException(Resource.ErrorExpirationDaysNeedsToBeSet);
            }

            StakedSuggestion stakedSuggestion = new StakedSuggestion(_expirationDays);
            Suggestion suggestion =
                dbServiceContext.Suggestions.FirstOrDefault(s => s.Id.ToString() == suggestionId.ToString());
            if (suggestion != null)
            {
                stakedSuggestion.IssueId = suggestion.IssueId;
                stakedSuggestion.UserId = userId;
            }

            dbServiceContext.StakedSuggestions.Add(stakedSuggestion);
        }

        /// <summary>
        /// Rollbacks the other stake for issue if already existing.
        /// </summary>
        /// <param name="dbServiceContext">The database service context.</param>
        /// <param name="walletService">The wallet service.</param>
        /// <param name="suggestionId">The suggestion identifier.</param>
        /// <param name="userId">The user identifier.</param>
        /// <param name="stakedSuggestionsForUser">The staked suggestions for user.</param>
        private static void RollbackOtherStakeForIssueIfAlreadyExisting(DbServiceContext dbServiceContext,
            WalletService walletService, Guid suggestionId, Guid userId,
            List<StakedSuggestion> stakedSuggestionsForUser)
        {
            Guid? issueId = stakedSuggestionsForUser
                .FirstOrDefault(s => s.SuggestionId.ToString() == suggestionId.ToString())?.IssueId;

            if (issueId != null)
            {
                StakedSuggestion existingStakedSuggestion = stakedSuggestionsForUser
                    .FirstOrDefault(s => s.IssueId.ToString() == issueId.ToString() &&
                                         suggestionId.ToString() != s.SuggestionId.ToString());

                if (existingStakedSuggestion != null)
                {
                    walletService.AddTransaction(dbServiceContext, userId, TransactionTypeNames.StakeSuggestionRollback,
                        suggestionId);

                    dbServiceContext.Remove(existingStakedSuggestion);
                }
            }
        }

        private static void ThrowExceptionIfAlreadyStaked(Guid suggestionId,
            List<StakedSuggestion> stakedSuggestionsForUser)
        {
            if (stakedSuggestionsForUser.Any(stakedSuggestion =>
                stakedSuggestion.SuggestionId.ToString() == suggestionId.ToString()))
            {
                throw new InvalidOperationException(Resource.ErrorSuggestionAlreadyStakedForUser);
            }
        }

        /// <summary>
        /// Rolls the back invalid staked suggestions.
        /// </summary>
        public void RollBackInvalidStakedSuggestions(DbServiceContext dbServiceContext)
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
                    Guid suggestionId = stakedSuggestion.SuggestionId;
                    walletService.AddTransaction(dbServiceContext, userId, TransactionTypeNames.StakeSuggestionRollback,
                        suggestionId);
                }
            }
        }

        /// <summary>
        /// Imports the specified data table.
        /// </summary>
        /// <param name="dataTable">The data table.</param>
        /// <returns>The number of imported records</returns>
        public int Import(DataTable dataTable)
        {
            DbServiceContext dbServiceContext = DatabaseInitializationService.GetDbServiceContext();

            using (dbServiceContext)
            {
                int count = dbServiceContext.StakedSuggestions.Count();

                if (count > 0)
                {
                    return 0;
                }

                int recordCount = 0;

                foreach (DataRow row in dataTable.Rows)
                {
                    StakedSuggestion stakedSuggestion = new StakedSuggestion
                    {
                        ImportId = Convert.ToInt32(row["ID"].ToString())
                    };

                    string suggestionId = row["SuggestionID"].ToString();

                    if (!string.IsNullOrEmpty(suggestionId))
                    {
                        Suggestion suggestion = dbServiceContext.Suggestions
                            .FirstOrDefault(s => s.ImportId == Convert.ToInt32(suggestionId));

                        if (suggestion != null)
                        {
                            stakedSuggestion.SuggestionId = suggestion.Id;
                            stakedSuggestion.IssueId = suggestion.IssueId;
                        }
                    }

                    string userId = row["UserID"].ToString();

                    if (!string.IsNullOrEmpty(userId))
                    {
                        User user = dbServiceContext.User
                            .FirstOrDefault(u => u.ImportId == Convert.ToInt32(userId));

                        if (user != null)
                        {
                            stakedSuggestion.UserId = user.Id;
                        }
                    }

                    if (stakedSuggestion.SuggestionId.ToString() != Guid.Empty.ToString() &&
                        stakedSuggestion.UserId.ToString() != Guid.Empty.ToString())
                    {
                        dbServiceContext.StakedSuggestions.Add(stakedSuggestion);
                        recordCount++;
                    }
                }

                if (recordCount > 0)
                {
                    dbServiceContext.SaveChanges();
                }

                return recordCount;
            }
        }
    }
}
