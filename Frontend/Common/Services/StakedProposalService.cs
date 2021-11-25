using System;
using System.Collections.Generic;
using System.Data;
using System.Linq;
using Common.Data;
using Common.Entities;
using Microsoft.EntityFrameworkCore;

namespace Common.Services
{
    public class StakedProposalService
    {
        /// <summary>
        /// The expiration days
        /// </summary>
        private readonly int _expirationDays;

        /// <summary>
        /// Initializes a new instance of the <see cref="StakedProposalService"/> class.
        /// </summary>
        public StakedProposalService()
        {
            _expirationDays = 0;
        }

        /// <summary>
        /// Initializes a new instance of the <see cref="StakedProposalService" /> class.
        /// </summary>
        /// <param name="expirationDays">The expiration days.</param>
        public StakedProposalService(int expirationDays)
        {
            _expirationDays = expirationDays;
        }

        /// <summary>
        /// Gets the staked suggestions for user.
        /// </summary>
        /// <param name="dbServiceContext">The database service context.</param>
        /// <param name="userId">The user identifier.</param>
        /// <returns>The staked suggestions for the user</returns>
        public List<StakedProposal> GetStakedSuggestionsForUser(DbServiceContext dbServiceContext, Guid userId)
        {
            RollBackInvalidStakedSuggestions(dbServiceContext);

            List<StakedProposal> stakedSuggestionsForUser = dbServiceContext.Users
                .FirstOrDefault(u => u.Id.ToString() == userId.ToString())?
                .StakedSuggestions ?? new List<StakedProposal>();

            return stakedSuggestionsForUser;
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

            List<StakedProposal> stakedSuggestionsForUser = GetStakedSuggestionsForUser(dbServiceContext, userId);

            ThrowExceptionIfAlreadyStaked(suggestionId, stakedSuggestionsForUser);

            WalletService walletService = new WalletService();

            RollbackOtherStakeForIssueIfAlreadyExisting(dbServiceContext, walletService, suggestionId, userId,
                stakedSuggestionsForUser);

            walletService.AddTransaction(dbServiceContext, userId, TransactionTypeNames.StakeProposal, suggestionId);

            AddStakedSuggestion(dbServiceContext, suggestionId, userId);

            dbServiceContext.SaveChanges();
        }

        /// <summary>
        /// Adds the stake suggestion.
        /// </summary>
        /// <param name="dbServiceContext">The database service context.</param>
        /// <param name="suggestionId">The suggestion identifier.</param>
        /// <param name="userId">The user identifier.</param>
        /// <exception cref="System.InvalidOperationException"></exception>
        private void AddStakedSuggestion(DbServiceContext dbServiceContext, Guid suggestionId, Guid userId)
        {
            if (_expirationDays == 0)
            {
                throw new InvalidOperationException(Resource.ErrorExpirationDaysNeedsToBeSet);
            }

            StakedProposal stakedSuggestion = new StakedProposal(_expirationDays);
            Proposal suggestion =
                dbServiceContext.Proposals.FirstOrDefault(s => s.Id.ToString() == suggestionId.ToString());

            if (suggestion != null)
            {
                stakedSuggestion.IssueId = suggestion.IssueId;
                stakedSuggestion.ProposalId = suggestion.Id;
                stakedSuggestion.UserId = userId;
            }

            dbServiceContext.StakedProposals.Add(stakedSuggestion);
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
            List<StakedProposal> stakedSuggestionsForUser)
        {
            Guid? issueId = stakedSuggestionsForUser
                .FirstOrDefault(s => s.ProposalId.ToString() == suggestionId.ToString())?.IssueId;

            if (issueId != null)
            {
                StakedProposal existingStakedSuggestion = stakedSuggestionsForUser
                    .FirstOrDefault(s => s.IssueId.ToString() == issueId.ToString() &&
                                         suggestionId.ToString() != s.ProposalId.ToString());

                if (existingStakedSuggestion != null)
                {
                    walletService.AddTransaction(dbServiceContext, userId, TransactionTypeNames.StakeProposalRollback,
                        suggestionId);

                    dbServiceContext.Remove(existingStakedSuggestion);
                }
            }
        }

        /// <summary>
        /// Throws the exception if already staked.
        /// </summary>
        /// <param name="suggestionId">The suggestion identifier.</param>
        /// <param name="stakedSuggestionsForUser">The staked suggestions for user.</param>
        /// <exception cref="System.InvalidOperationException">ErrorSuggestionAlreadyStakedForUser</exception>
        private static void ThrowExceptionIfAlreadyStaked(Guid suggestionId,
            List<StakedProposal> stakedSuggestionsForUser)
        {
            if (stakedSuggestionsForUser.Any(stakedSuggestion =>
                stakedSuggestion.ProposalId.ToString() == suggestionId.ToString()))
            {
                throw new InvalidOperationException(Resource.ErrorProposalAlreadyStakedForUser);
            }
        }

        /// <summary>
        /// Rolls the back invalid staked suggestions.
        /// </summary>
        public void RollBackInvalidStakedSuggestions(DbServiceContext dbServiceContext)
        {
            List<StakedProposal> invalidStakedSuggestions = dbServiceContext.StakedProposals
                .Where(s => s.CreateDate.AddDays(s.ExpirationDays) < DateTime.Now).ToList();

            if (invalidStakedSuggestions.Count == 0)
            {
                return;
            }

            WalletService walletService = new WalletService();

            foreach (User user in dbServiceContext.Users.Include(u => u.StakedSuggestions).ToList())
            {
                foreach (StakedProposal stakedSuggestion in invalidStakedSuggestions)
                {
                    if (!user.StakedSuggestions.Contains(stakedSuggestion))
                    {
                        continue;
                    }

                    Guid userId = user.Id;
                    Guid suggestionId = stakedSuggestion.ProposalId;
                    walletService.AddTransaction(dbServiceContext, userId, TransactionTypeNames.StakeProposalRollback,
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
                int count = dbServiceContext.StakedProposals.Count();

                if (count > 0)
                {
                    return 0;
                }

                int recordCount = 0;

                foreach (DataRow row in dataTable.Rows)
                {
                    StakedProposal stakedProposal = new StakedProposal
                    {
                        ImportId = Convert.ToInt32(row["ID"].ToString())
                    };

                    string suggestionId = row["ProposalID"].ToString();

                    if (!string.IsNullOrEmpty(suggestionId))
                    {
                        Proposal suggestion = dbServiceContext.Proposals
                            .FirstOrDefault(s => s.ImportId == Convert.ToInt32(suggestionId));

                        if (suggestion != null)
                        {
                            stakedProposal.ProposalId = suggestion.Id;
                            stakedProposal.IssueId = suggestion.IssueId;
                        }
                    }

                    string userId = row["UserID"].ToString();

                    if (!string.IsNullOrEmpty(userId))
                    {
                        User user = dbServiceContext.Users
                            .FirstOrDefault(u => u.ImportId == Convert.ToInt32(userId));

                        if (user != null)
                        {
                            stakedProposal.UserId = user.Id;
                        }
                    }

                    stakedProposal.ExpirationDays = 30;

                    if (stakedProposal.ProposalId.ToString() != Guid.Empty.ToString() &&
                        stakedProposal.UserId.ToString() != Guid.Empty.ToString())
                    {
                        dbServiceContext.StakedProposals.Add(stakedProposal);
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
