using System;
using System.Collections.Generic;
using System.Data;
using System.Linq;
using Common.Data;
using Common.Entities;
using Common.Interfaces;
using Microsoft.EntityFrameworkCore;

namespace Common.Services
{
    /// <summary>
    /// Implementation of the issue service
    /// </summary>
    /// <seealso cref="Common.Interfaces.IIssueService" />
    public class IssueService : IIssueService
    {
        /// <summary>
        /// Gets all issues.
        /// </summary>
        /// <returns>
        /// Gets all issues
        /// </returns>
        public List<Issue> GetAllIssues(bool includeSuggestions = false)
        {
            DbServiceContext dbServiceContext = DatabaseInitializationService.GetDbServiceContext();

            using (dbServiceContext)
            {
                List<Issue> issues;

                if (!includeSuggestions)
                {
                    issues = dbServiceContext.Issues.ToList();
                }
                else
                {
                    issues = dbServiceContext.Issues
                        .Include(i => i.Suggestions)
                        .ToList();

                    foreach (Issue issue in issues)
                    {
                        SuggestionService.UpdateStakes(issue.Suggestions);
                    }
                }

                return issues;
            }
        }

        /// <summary>
        /// Gets all valid issues.
        /// </summary>
        /// <param name="includeSuggestions"></param>
        /// <param name="onlyStaked"></param>
        /// <returns>
        /// All valid issues that contain at least one stake
        /// </returns>
        public List<Issue> GetAllValidIssues(bool includeSuggestions = false, bool onlyStaked = false)
        {
            // TODO: DueDate can be null

            DbServiceContext dbServiceContext = DatabaseInitializationService.GetDbServiceContext();

            using (dbServiceContext)
            {
                if (!includeSuggestions)
                {
                    return dbServiceContext.Issues
                        .Where(issue => issue.DueDate >= DateTime.Now)
                        .ToList();
                }

                List<Issue> issues = dbServiceContext.Issues
                    .Include(i => i.Suggestions)
                    .Where(issue => issue.DueDate >= DateTime.Now)
                    .ToList();

                foreach (Issue issue in issues)
                {
                    SuggestionService.UpdateStakes(issue.Suggestions);
                }

                return !onlyStaked ? 
                    issues : issues.Where(issue => issue.Suggestions.Any(suggestion => suggestion.IsStaked)).ToList();
            }
        }

        /// <summary>
        /// Gets the top staked issues.
        /// </summary>
        /// <param name="limit">The limit.</param>
        /// <returns>
        /// The top staked issues depending on the limit
        /// </returns>
        public List<Issue> GetTopStakedIssues(int limit)
        {
            List<Issue> issues = new List<Issue>(GetAllValidIssues(true, true));

            if (limit <= issues.Count)
            {
                return issues;
            }

            return issues.OrderByDescending(i => i.GetTotalStakeCount()).Take(limit).ToList();
        }

        /// <summary>
        /// Gets the top staked issues percent.
        /// </summary>
        /// <param name="percentage">The percentage.</param>
        /// <param name="limitNumber">The limit number.</param>
        /// <returns>
        /// The top stacked issues depending on a percentage
        /// </returns>
        public List<Issue> GetTopStakedIssuesPercentage(decimal percentage, int limitNumber = 0)
        {
            List<Issue> issues = new List<Issue>(GetAllValidIssues(true, true));

            if (percentage >= 100)
            {
                return issues;
            }

            decimal count = Convert.ToDecimal(issues.Count);

            int limit = Convert.ToInt32(Math.Round(percentage / 100 * count));

            if (limitNumber > 0 && limit > limitNumber)
            {
                limit = limitNumber;
            }

            return GetTopStakedIssues(limit);
        }

        /// <summary>
        /// Gets the issues by tags.
        /// </summary>
        /// <param name="tags">The tags.</param>
        /// <returns>
        /// All the issues that contain to at least one of the given tags
        /// </returns>
        public List<Issue> GetIssuesByTags(string tags)
        {
            List<string> tagsList = new List<string>(Issue.GetTags(tags));

            List<Issue> issues = new List<Issue>();

            foreach (Issue issue in GetAllValidIssues(true, true))
            {
                if (tagsList.Any(tag => issue.HasTag(tag)))
                {
                    issues.Add(issue);
                }
            }

            return issues;
        }

        /// <summary>
        /// Gets the top staked issues by tags.
        /// </summary>
        /// <param name="tags">The tags.</param>
        /// <param name="limit">The limit.</param>
        /// <returns>
        /// The top staked issues by tags
        /// </returns>
        public List<Issue> GetTopStakedIssuesByTags(string tags, int limit)
        {
            List<Issue> issues = new List<Issue>(GetIssuesByTags(tags));

            if (limit == 0)
            {
                return issues;
            }

            return issues.OrderByDescending(i => i.GetTotalStakeCount()).Take(limit).ToList();
        }

        /// <summary>
        /// Gets the top stakes issues percentage by tags.
        /// </summary>
        /// <param name="tags">The tags.</param>
        /// <param name="percentage">The percentage.</param>
        /// <param name="limitNumber">The limit number.</param>
        /// <returns>
        /// The top staked issues in percent by tags
        /// </returns>
        public List<Issue> GetTopStakesIssuesPercentageByTags(string tags, decimal percentage = 100, int limitNumber = 0)
        {
            List<Issue> issues = new List<Issue>(GetIssuesByTags(tags));

            if (percentage >= 100)
            {
                return issues;
            }

            decimal count = Convert.ToDecimal(issues.Count);

            int limit = Convert.ToInt32(Math.Round(percentage / 100 * count));

            if (limitNumber > 0 && limit > limitNumber)
            {
                limit = limitNumber;
            }

            return GetTopStakedIssues(limit);
        }

        /// <summary>
        /// Adds the issue.
        /// </summary>
        /// <param name="issue">The issue.</param>
        /// <param name="userId">The user identifier.</param>
        /// <returns>
        /// The guid of the issue added
        /// </returns>
        /// <exception cref="System.InvalidOperationException">
        /// Will be thrown if user is not found
        /// </exception>
        public Guid AddIssue(Issue issue, Guid userId)
        {
            (bool valid, string errorMessage) = IsValid(issue);

            if (!valid)
            {
                throw new InvalidOperationException(errorMessage);
            }

            TransactionTypeService transactionTypeService = new TransactionTypeService();
            TransactionType transactionType = transactionTypeService.GetTransactionType(TransactionTypeNames.AddIssue);

            UserService userService = new UserService();
            User user = userService.GetUserById(userId);

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
                Balance = transactionType.Fee,
                CreateDate = DateTime.Now,
                TransactionType = transactionType
            };

            wallet.AddTransaction(walletTransaction);

            issue.Suggestions ??= new List<Suggestion>();

            DbServiceContext dbServiceContext = DatabaseInitializationService.GetDbServiceContext();

            using (dbServiceContext)
            {
                dbServiceContext.Issues.Add(issue);
                dbServiceContext.SaveChanges();
                return issue.Id;
            }
        }

        /// <summary>
        /// Gets the issue.
        /// </summary>
        /// <param name="issueId">The issue unique identifier.</param>
        /// <param name="includeSuggestions">if set to <c>true</c> [with suggestions].</param>
        /// <returns>
        /// The issue for the given issue guid
        /// </returns>
        public Issue GetIssue(Guid issueId, bool includeSuggestions = false)
        {
            DbServiceContext dbServiceContext = DatabaseInitializationService.GetDbServiceContext();

            using (dbServiceContext)
            {
                if (!includeSuggestions)
                {
                    return dbServiceContext.Issues
                        .FirstOrDefault(i => i.Id.ToString() == issueId.ToString());
                }

                Issue issue = dbServiceContext.Issues
                    .Include(i => i.Suggestions)
                    .FirstOrDefault(i => i.Id.ToString() == issueId.ToString());

                if (issue != null)
                {
                    SuggestionService.UpdateStakes(issue.Suggestions);
                }

                return issue;
            }
        }

        /// <summary>
        /// Imports the specified data table.
        /// </summary>
        /// <param name="dataTable">The data table.</param>
        /// <returns>
        /// The number of imported records
        /// </returns>
        public int Import(DataTable dataTable)
        {
            DbServiceContext dbServiceContext = DatabaseInitializationService.GetDbServiceContext();

            using (dbServiceContext)
            {
                int count = dbServiceContext.Issues.Count();

                if (count > 0)
                {
                    return 0;
                }

                int recordCount = 0;

                foreach (DataRow row in dataTable.Rows)
                {
                    Issue issue = new Issue
                    {
                        ImportId = Convert.ToInt32(row["ID"].ToString()),
                        Tags = row["Tags"].ToString(),
                        Description = row["Description"].ToString(),
                        Title = row["Title"].ToString(),
                        Suggestions = new List<Suggestion>()
                    };

                    string value = row["DueDateDays"].ToString();

                    if (value != null)
                    {
                        issue.DueDate = DateTime.Now.Date.AddDays(Convert.ToInt32(value));
                    }

                    dbServiceContext.Issues.AddRange(issue);

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
        /// Returns true if issue is valid.
        /// </summary>
        /// <param name="issue">The issue.</param>
        /// <returns>
        ///   <c>true</c> if this instance is valid; otherwise, <c>false</c>.
        /// </returns>
        private static (bool valid, string errorMessage) IsValid(Issue issue)
        {
            string errorMessage = string.Empty;

            if (string.IsNullOrEmpty(issue.Tags))
            {
                errorMessage = Resource.ErrorTagsAreRequired;
                return (false, errorMessage);
            }

            if (string.IsNullOrEmpty(issue.Title))
            {
                errorMessage = Resource.ErrorTitleIsRequired;
                return (false, errorMessage);
            }

            if (string.IsNullOrEmpty(issue.Description))
            {
                errorMessage = Resource.ErrorDescriptionIsRequired;
                return (false, errorMessage);
            }

            if (!string.IsNullOrEmpty(issue.Tags) && issue.Tags.Length < 5)
            {
                errorMessage = Resource.ErrorTagsNotLongEnough;
                return (false, errorMessage);
            }

            if (!string.IsNullOrEmpty(issue.Title) && issue.Title.Length < 5)
            {
                errorMessage = Resource.ErrorTitleNotLongEnough;
                return (false, errorMessage);
            }

            if (!string.IsNullOrEmpty(issue.Description) && issue.Description.Length < 5)
            {
                errorMessage = Resource.ErrorDescriptionNotLongEnough;
                return (false, errorMessage);
            }

            double? differenceDays = issue.DueDate?.Subtract(DateTime.Now).TotalDays;

            if (differenceDays is < 5)
            {
                errorMessage = Resource.ErrorDueDateToShort;
                return (false, errorMessage);
            }

            return (true, errorMessage);
        }
    }
}
