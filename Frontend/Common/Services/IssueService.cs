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
        public List<Issue> GetAllIssues()
        {
            DbServiceContext dbServiceContext = DatabaseInitializationService.GetDbServiceContext();

            using (dbServiceContext)
            {
                return dbServiceContext.Issues
                    .Include(i => i.Suggestions)
                    .ToList();
            }
        }

        /// <summary>
        /// Gets all valid issues.
        /// </summary>
        /// <param name="onlyStaked"></param>
        /// <returns>
        /// All valid issues that contain at least one stake
        /// </returns>
        public List<Issue> GetAllValidIssues(bool onlyStaked = false)
        {
            DbServiceContext dbServiceContext = DatabaseInitializationService.GetDbServiceContext();

            using (dbServiceContext)
            {
                List<Issue> allIssues = dbServiceContext.Issues
                    .Where(issue => issue.DueDate >= DateTime.Now).ToList();

                return !onlyStaked ? allIssues : allIssues.Where(issue => issue.Suggestions.Any(suggestion => suggestion.IsStaked)).ToList();
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
            List<Issue> issues = new List<Issue>(GetAllValidIssues());

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
        /// <returns>
        /// The top stacked issues depending on a percentage
        /// </returns>
        public List<Issue> GetTopStakedIssuesPercentage(decimal percentage)
        {
            List<Issue> issues = new List<Issue>(GetAllValidIssues());

            if (percentage >= 100)
            {
                return issues;
            }

            decimal count = Convert.ToDecimal(issues.Count);

            int limit = Convert.ToInt32(Math.Round(percentage / 100 * count));

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

            foreach (Issue issue in GetAllValidIssues())
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

        public List<Issue> GetTopStakesIssuesPercentageByTags(string tags, decimal percentage = 100)
        {
            List<Issue> issues = new List<Issue>(GetIssuesByTags(tags));

            if (percentage >= 100)
            {
                return issues;
            }

            decimal count = Convert.ToDecimal(issues.Count);

            int limit = Convert.ToInt32(Math.Round(percentage / 100 * count));

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
        /// <returns>
        /// The issue for the given issue guid
        /// </returns>
        public Issue GetIssue(Guid issueId)
        {
            DbServiceContext dbServiceContext = DatabaseInitializationService.GetDbServiceContext();

            using (dbServiceContext)
            {
                return dbServiceContext.Issues.FirstOrDefault(i => i.Id.ToString() == issueId.ToString());
            }
        }

        public int Import(DataTable issues)
        {
            throw new NotImplementedException();
        }
    }
}
