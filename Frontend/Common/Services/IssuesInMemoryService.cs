using System;
using System.Collections.Generic;
using System.Data;
using System.Linq;
using Common.Entities;
using Common.Interfaces;

namespace Common.Services
{
    /// <summary>
    /// Implementation of an abstract issue service
    /// </summary>
    public class IssuesInMemoryService : IIssuesService
    {
        /// <summary>
        /// All issues
        /// </summary>
        private readonly List<Issue> _allIssues;

        /// <summary>
        /// The transaction types
        /// </summary>
        private readonly List<TransactionType> _transactionTypes;

        /// <summary>
        /// Initializes a new instance of the <see cref="IssuesInMemoryService" /> class.
        /// </summary>
        /// <param name="allIssues">All issues.</param>
        /// <param name="transactionTypes">The transaction types.</param>
        public IssuesInMemoryService(List<Issue> allIssues, List<TransactionType> transactionTypes)
        {
            _allIssues = allIssues;
            _transactionTypes = transactionTypes;
        }

        /// <summary>
        /// Gets all issues.
        /// </summary>
        /// <returns>Gets all issues</returns>
        public List<Issue> GetAllIssues()
        {
            return _allIssues;
        }

        /// <summary>
        /// Gets all valid issues.
        /// </summary>
        /// <returns>All valid issues that contain at least one stake</returns>
        public IEnumerable<Issue> GetAllValidIssues(bool onlyStaked = false)
        {
            IEnumerable<Issue> allIssues = GetAllIssues()
                .Where(issue => issue.DueDate == null || issue.DueDate >= DateTime.Now);

            return !onlyStaked ? allIssues : allIssues.Where(issue => issue.Suggestions.Any(suggestion => suggestion.IsStaked)).ToList();
        }

        /// <summary>
        /// Gets the top staked issues.
        /// </summary>
        /// <param name="limit">The limit.</param>
        /// <returns>The top staked issues depending on the limit</returns>
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
        /// <returns>All the issues that contain to at least one of the given tags</returns>
        public IEnumerable<Issue> GetIssuesByTags(string tags)
        {
            List<string> tagsList = new List<string>(Issue.GetTags(tags));

            foreach (Issue issue in GetAllValidIssues())
            {
                if (tagsList.Any(tag => issue.HasTag(tag)))
                {
                    yield return issue;
                }
            }
        }

        /// <summary>
        /// Gets the top staked issues by tags.
        /// </summary>
        /// <param name="tags">The tags.</param>
        /// <param name="limit">The limit.</param>
        /// <returns>The top staked issues by tags</returns>
        public IEnumerable<Issue> GetTopStakedIssuesByTags(string tags, int limit)
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
        /// <returns>The top staked issues in percent by tags</returns>
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
        /// <param name="user">The user.</param>
        /// <returns>
        /// The guid of the issue added
        /// </returns>
        /// <exception cref="System.InvalidOperationException"></exception>
        public Guid AddIssue(Issue issue, User user)
        {
            TransactionTypeService transactionTypeBaseService = new TransactionTypeService();
            TransactionType transactionType = transactionTypeBaseService.GetTransactionType(TransactionTypeNames.AddIssue);

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

            _allIssues.Add(issue);

            return issue.Id;
        }

        /// <summary>
        /// Gets the issue.
        /// </summary>
        /// <param name="issueId">The issue identifier.</param>
        /// <returns>The issue for the given id or null if not found</returns>
        public Issue GetIssue(Guid issueId)
        {
            return _allIssues.FirstOrDefault(i => i.Id.ToString() == issueId.ToString());
        }

        public int Import(DataTable issues)
        {
            throw new NotImplementedException();
        }
    }
}
