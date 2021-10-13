using System;
using System.Collections.Generic;
using System.Data;
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
        private readonly decimal _topStakedSuggestionsPercent;

        private readonly decimal _topVotedSuggestionsPercent;

        /// <summary>
        /// Initializes a new instance of the <see cref="SuggestionService"/> class.
        /// </summary>
        public SuggestionService()
        {
            _topStakedSuggestionsPercent = 0;
            _topVotedSuggestionsPercent = 0;
        }

        /// <summary>
        /// Initializes a new instance of the <see cref="SuggestionService" /> class.
        /// </summary>
        /// <param name="topStakedSuggestionsPercent">The top staked suggestions percent.</param>
        /// <param name="topVotedSuggestionsPercent">The top voted suggestions percent.</param>
        public SuggestionService(decimal topStakedSuggestionsPercent, decimal topVotedSuggestionsPercent)
        {
            _topStakedSuggestionsPercent = topStakedSuggestionsPercent;
            _topVotedSuggestionsPercent = topVotedSuggestionsPercent;
        }

        /// <summary>
        /// Gets the by identifier.
        /// </summary>
        /// <param name="dbServiceContext">The database service context.</param>
        /// <param name="id">The identifier.</param>
        /// <param name="userId">The user identifier.</param>
        /// <returns>
        /// Gets the suggestions for the given id
        /// </returns>
        /// <exception cref="System.InvalidOperationException"></exception>
        public List<Suggestion> GetByIssueId(DbServiceContext dbServiceContext, string id, string userId)
        {
            if (_topStakedSuggestionsPercent == 0)
            {
                throw new InvalidOperationException(Resource.ErrorTopStakePercentNeedsToBeSet);
            }

            Issue issue = dbServiceContext.Issues
                .Include(i => i.Suggestions)
                .FirstOrDefault(i => i.Id.ToString() == id);

            if (issue == null)
            {
                return null;
            }

            issue.Suggestions ??= new List<Suggestion>();

            UpdateStakes(dbServiceContext, issue.Suggestions);

            SetTopStaked(issue.Suggestions);

            SetHasMyStake(dbServiceContext, issue, userId);

            UpdateVotes(dbServiceContext, issue.Suggestions);

            SetTopVoted(issue.Suggestions);

            SetMyVote(dbServiceContext, issue, userId);

            return issue.Suggestions.OrderByDescending(s => s.IsTopStaked).ThenBy(s => s.CreateDate).ToList();
        }

        /// <summary>
        /// Gets the by suggestion identifier.
        /// </summary>
        /// <param name="dbServiceContext">The database service context.</param>
        /// <param name="id">The identifier.</param>
        /// <param name="userId">The user identifier.</param>
        /// <returns></returns>
        /// <exception cref="System.InvalidOperationException"></exception>
        public Suggestion GetBySuggestionId(DbServiceContext dbServiceContext, string id, string userId = null)
        {
            if (_topStakedSuggestionsPercent == 0)
            {
                throw new InvalidOperationException(Resource.ErrorTopStakePercentNeedsToBeSet);
            }

            Suggestion suggestion = dbServiceContext.Suggestions
                .FirstOrDefault(s => s.Id.ToString() == id);

            if (suggestion == null)
            {
                return null;
            }

            UpdateStakes(dbServiceContext, new List<Suggestion> { suggestion });

            SetTopStaked(new List<Suggestion> { suggestion });

            if (!string.IsNullOrEmpty(userId))
            {
                SetHasMyStake(dbServiceContext, suggestion, userId);
            }

            UpdateVotes(dbServiceContext, new List<Suggestion> { suggestion });

            SetTopVoted(new List<Suggestion> { suggestion });

            if (!string.IsNullOrEmpty(userId))
            {
                SetMyVote(dbServiceContext, suggestion, userId);
            }

            return suggestion;
        }

        /// <summary>
        /// Sets the has my stake.
        /// </summary>
        /// <param name="dbServiceContext">The database service context.</param>
        /// <param name="issue">The issue.</param>
        /// <param name="userId">The user identifier.</param>
        public static void SetHasMyStake(DbServiceContext dbServiceContext, Issue issue, string userId)
        {
            if (string.IsNullOrEmpty(userId))
            {
                return;
            }

            StakedSuggestion stakedSuggestion = dbServiceContext.StakedSuggestions
                .FirstOrDefault(s => s.IssueId.ToString() == issue.Id.ToString() && s.UserId.ToString() == userId);

            if (stakedSuggestion != null)
            {
                foreach (var suggestion in issue.Suggestions
                    .Where(suggestion => suggestion.Id.ToString() == stakedSuggestion.SuggestionId.ToString()))
                {
                    suggestion.HasMyStake = true;
                    break;
                }
            }
        }

        /// <summary>
        /// Sets the has my stake.
        /// </summary>
        /// <param name="dbServiceContext">The database service context.</param>
        /// <param name="suggestion">The suggestion.</param>
        /// <param name="userId">The user identifier.</param>
        public static void SetHasMyStake(DbServiceContext dbServiceContext, Suggestion suggestion, string userId)
        {
            if (string.IsNullOrEmpty(userId))
            {
                return;
            }

            StakedSuggestion stakedSuggestion = dbServiceContext.StakedSuggestions
                .FirstOrDefault(s => s.SuggestionId.ToString() == suggestion.Id.ToString() && s.UserId.ToString() == userId);

            if (stakedSuggestion != null)
            {
                suggestion.HasMyStake = true;
            }
        }

        /// <summary>
        /// Updates the stakes.
        /// </summary>
        /// <param name="dbServiceContext">The database service context.</param>
        /// <param name="suggestions">The suggestions.</param>
        public static void UpdateStakes(DbServiceContext dbServiceContext, List<Suggestion> suggestions)
        {
            foreach (Suggestion suggestion in suggestions)
            {
                int count = dbServiceContext.StakedSuggestions
                    .Where(s => s.SuggestionId.ToString() == suggestion.Id.ToString())
                    .ToList().Count;

                suggestion.StakeCount = count;
            }
        }

        /// <summary>
        /// Sets my vote.
        /// </summary>
        /// <param name="dbServiceContext">The database service context.</param>
        /// <param name="issue">The issue.</param>
        /// <param name="userId">The user identifier.</param>
        public static void SetMyVote(DbServiceContext dbServiceContext, Issue issue, string userId)
        {
            if (string.IsNullOrEmpty(userId))
            {
                return;
            }

            List<Vote> votes = dbServiceContext.Votes
                .Include(v => v.Suggestion)
                .Where(v => v.UserId.ToString() == userId && v.IssueId.ToString() == issue.Id.ToString())
                .ToList();

            if (votes.Count == 0)
            {
                return;
            }

            foreach (Suggestion suggestion in issue.Suggestions)
            {
                Vote vote = votes.FirstOrDefault(v => v.SuggestionId.ToString() == suggestion.Id.ToString());

                if (vote != null)
                {
                    suggestion.MyVote = vote.Value;
                }
            }
        }

        /// <summary>
        /// Sets my vote.
        /// </summary>
        /// <param name="dbServiceContext">The database service context.</param>
        /// <param name="suggestion">The suggestion.</param>
        /// <param name="userId">The user identifier.</param>
        public static void SetMyVote(DbServiceContext dbServiceContext, Suggestion suggestion, string userId)
        {
            if (string.IsNullOrEmpty(userId))
            {
                return;
            }

            var vote = dbServiceContext.Votes
                .Include(v => v.Suggestion)
                .FirstOrDefault(v => v.UserId.ToString() == userId && v.SuggestionId.ToString() == suggestion.Id.ToString());

            if (vote == null)
            {
                return;
            }

            suggestion.MyVote = vote.Value;
        }

        /// <summary>
        /// Updates the votes.
        /// </summary>
        /// <param name="dbServiceContext">The database service context.</param>
        /// <param name="suggestions">The suggestions.</param>
        public static void UpdateVotes(DbServiceContext dbServiceContext, List<Suggestion> suggestions)
        {
            foreach (Suggestion suggestion in suggestions)
            {
                int count = dbServiceContext.Votes
                    .Where(v => v.SuggestionId.ToString() == suggestion.Id.ToString())
                    .ToList().Count;

                suggestion.VoteCount = count;
            }
        }

        /// <summary>
        /// Adds the specified database service context.
        /// </summary>
        /// <param name="dbServiceContext">The database service context.</param>
        /// <param name="suggestionSubmission">The suggestion submission.</param>
        /// <returns></returns>
        /// <exception cref="System.InvalidOperationException">
        /// </exception>
        public Guid Add(DbServiceContext dbServiceContext, SuggestionSubmission suggestionSubmission)
        {
            Suggestion suggestion = suggestionSubmission.ToSuggestion();

            (bool valid, string errorMessage) = IsValid(suggestion);

            if (!valid)
            {
                throw new InvalidOperationException(errorMessage);
            }

            string issueId = suggestionSubmission.IssueId.ToString();

            bool suggestionWithSameTitleAlreadyExists = dbServiceContext.Suggestions
                .FirstOrDefault(s => s.IssueId.ToString() == issueId &&
                                     string.Equals(s.Title, suggestionSubmission.Title, StringComparison.OrdinalIgnoreCase)) != null;

            if (suggestionWithSameTitleAlreadyExists)
            {
                throw new InvalidOperationException(Resource.ErrorSuggestionWithSameTitleAlreadyExists);
            }

            Guid userId = suggestionSubmission.UserId;

            TransactionTypeService transactionTypeService = new TransactionTypeService();
            TransactionType transactionType = transactionTypeService.GetTransactionType(dbServiceContext, TransactionTypeNames.AddSuggestion);

            UserService userService = new UserService();
            User user = userService.GetUserById(dbServiceContext, userId);

            if (user == null)
            {
                throw new InvalidOperationException(string.Format(Resource.ErrorUserIdNotFound, userId));
            }

            Wallet wallet = user.Wallet;

            if (!wallet.HasEnoughFunding(transactionType.Fee))
            {
                throw new InvalidOperationException(Resource.ErrorNotEnoughFounding);
            }

            WalletTransaction walletTransaction = new WalletTransaction
            {
                // transaction fee must be negative for cost
                WalletId = wallet.Id,
                Balance = transactionType.Fee,
                CreateDate = DateTime.Now,
                TransactionType = transactionType,
                TransactionId = suggestion.Id
            };

            wallet.TotalBalance += walletTransaction.Balance;
            dbServiceContext.WalletTransactions.Add(walletTransaction);


            dbServiceContext.Suggestions.Add(suggestion);
            dbServiceContext.SaveChanges();

            return suggestion.Id;
        }

        /// <summary>
        /// Updates the specified database service context.
        /// </summary>
        /// <param name="dbServiceContext">The database service context.</param>
        /// <param name="suggestionSubmission">The suggestion submission.</param>
        /// <exception cref="System.InvalidOperationException">
        /// </exception>
        public void Update(DbServiceContext dbServiceContext, SuggestionSubmission suggestionSubmission)
        {
            Suggestion suggestionToUpdate = dbServiceContext.Suggestions
                .FirstOrDefault(s => s.Id.ToString() == suggestionSubmission.Id.ToString());

            if (suggestionToUpdate == null)
            {
                throw new InvalidOperationException(Resource.ErrorIssueNotFound);
            }

            if (!suggestionToUpdate.CanEdit(suggestionSubmission.UserId))
            {
                throw new InvalidOperationException(Resource.IssueCannotBeEditedAnymore);
            }

            suggestionToUpdate.Description = suggestionSubmission.Description;
            suggestionToUpdate.Title = suggestionSubmission.Title;

            (bool valid, string errorMessage) = IsValid(suggestionSubmission.ToSuggestion());

            if (!valid)
            {
                throw new InvalidOperationException(errorMessage);
            }

            dbServiceContext.Suggestions.Update(suggestionToUpdate);
            dbServiceContext.SaveChanges();
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
                int count = dbServiceContext.Suggestions.Count();

                if (count > 0)
                {
                    return 0;
                }

                int recordCount = 0;

                foreach (DataRow row in dataTable.Rows)
                {
                    Suggestion suggestion = new Suggestion
                    {
                        ImportId = Convert.ToInt32(row["ID"].ToString()),
                        Description = row["Description"].ToString(),
                        Title = row["Title"].ToString(),
                    };

                    int importIssueId = Convert.ToInt32(row["IssueId"].ToString());

                    Guid? issueId = dbServiceContext.Issues.FirstOrDefault(i => i.ImportId == importIssueId)?.Id;

                    if (issueId != null)
                    {
                        suggestion.IssueId = (Guid)issueId;
                    }

                    string userId = row["CreatorUserID"].ToString();

                    if (!string.IsNullOrEmpty(userId))
                    {
                        User user = dbServiceContext.User
                            .FirstOrDefault(u => u.ImportId == Convert.ToInt32(userId));

                        if (user != null)
                        {
                            suggestion.CreatorUserId = user.Id;
                        }
                    }

                    dbServiceContext.Suggestions.Add(suggestion);

                    recordCount++;
                }

                if (recordCount > 0)
                {
                    dbServiceContext.SaveChanges();
                }

                return recordCount;
            }
        }

        /// <summary>
        /// Sets the top staked.
        /// </summary>
        /// <param name="suggestions">The suggestions.</param>
        private void SetTopStaked(List<Suggestion> suggestions)
        {
            int topStakedIssuesCount = (int)Math.Round(suggestions.Count * _topStakedSuggestionsPercent / 100, 0);

            List<Suggestion> topStakedSuggestions = suggestions
                .OrderByDescending(i => i.StakeCount)
                .Take(topStakedIssuesCount)
                .ToList();

            foreach (var suggestion in topStakedSuggestions)
            {
                suggestion.IsTopStaked = true;
            }
        }

        /// <summary>
        /// Sets the top voted.
        /// </summary>
        /// <param name="suggestions">The suggestions.</param>
        private void SetTopVoted(List<Suggestion> suggestions)
        {
            int topVotedCount = (int)Math.Round(suggestions.Count * _topStakedSuggestionsPercent / 100, 0);

            List<Suggestion> topVotedSuggestions = suggestions
                .OrderByDescending(i => i.VoteCount)
                .Take(topVotedCount)
                .ToList();

            foreach (var suggestion in topVotedSuggestions)
            {
                suggestion.IsTopVoted = true;
            }
        }

        /// <summary>
        /// Returns true if ... is valid.
        /// </summary>
        /// <param name="suggestion">The suggestion.</param>
        /// <returns>
        ///   <c>true</c> if this instance is valid; otherwise, <c>false</c>.
        /// </returns>
        private (bool, string) IsValid(Suggestion suggestion)
        {
            string errorMessage = string.Empty;

            if (string.IsNullOrEmpty(suggestion.Title))
            {
                errorMessage = Resource.ErrorTitleIsRequired;
                return (false, errorMessage);
            }

            if (string.IsNullOrEmpty(suggestion.Description))
            {
                errorMessage = Resource.ErrorDescriptionIsRequired;
                return (false, errorMessage);
            }

            if (!string.IsNullOrEmpty(suggestion.Title) && suggestion.Title.Length < 5)
            {
                errorMessage = Resource.ErrorTitleNotLongEnough;
                return (false, errorMessage);
            }

            if (!string.IsNullOrEmpty(suggestion.Description) && suggestion.Description.Length < 5)
            {
                errorMessage = Resource.ErrorDescriptionNotLongEnough;
                return (false, errorMessage);
            }

            return (true, errorMessage);
        }
    }
}
