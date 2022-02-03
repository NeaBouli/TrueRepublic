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
    /// Implementation of the issue service
    /// </summary>
    public class IssueService
    {
        /// <summary>
        /// The top staked issues percent
        /// </summary>
        private readonly decimal _topStakedIssuesPercent;

        /// <summary>
        /// Initializes a new instance of the <see cref="IssueService"/> class.
        /// </summary>
        public IssueService()
        {
            _topStakedIssuesPercent = 0;
        }

        /// <summary>
        /// Initializes a new instance of the <see cref="IssueService"/> class.
        /// </summary>
        /// <param name="topStakedIssuesPercent">The top staked issues percent.</param>
        public IssueService(decimal topStakedIssuesPercent)
        {
            _topStakedIssuesPercent = topStakedIssuesPercent;
        }

        /// <summary>
        /// Gets all issues.
        /// </summary>
        /// <param name="dbServiceContext">The database service context.</param>
        /// <param name="paginatedList">The paginated list.</param>
        /// <param name="userId">The user identifier.</param>
        /// <returns>
        /// Gets all issues
        /// </returns>
        /// <exception cref="System.InvalidOperationException"></exception>
        public List<Issue> GetAll(DbServiceContext dbServiceContext, PaginatedList paginatedList = null, string userId = null)
        {
            List<Issue> issues = dbServiceContext.Issues
                .Include(i => i.Proposals)
                .Where(i => i.DueDate == null || (DateTime)i.DueDate >= DateTime.Now)
                .ToList();

            foreach (Issue issue in issues)
            {
                ProposalService.UpdateStakes(dbServiceContext, issue.Proposals);
                ProposalService.SetHasMyStake(dbServiceContext, issue, userId);

                issue.HasMyStake = issue.Proposals.Any(proposal => proposal.HasMyStake);
                issue.TotalStakeCount = issue.Proposals.Sum(proposal => proposal.StakeCount);
                issue.TotalVoteCount = dbServiceContext.Votes
                    .Where(vote => vote.IssueId.ToString() == issue.Id.ToString())
                    .ToList().Count;
            }

            if (_topStakedIssuesPercent > 0)
            {
                SetTopStaked(issues);
            }

            List<Issue> issuesProcessed = issues.OrderByDescending(i => i.TotalStakeCount)
                .ThenBy(i => i.CreateDate).ToList();

            issuesProcessed = ProcessPaginatedList(paginatedList, issuesProcessed);

            return issuesProcessed;
        }

        /// <summary>
        /// Gets the top staked issues percent.
        /// </summary>
        /// <param name="dbServiceContext">The database service context.</param>
        /// <param name="paginatedList">The paginated list.</param>
        /// <param name="userId">The user identifier.</param>
        /// <returns>
        /// The top stacked issues depending on a percentage
        /// </returns>
        /// <exception cref="System.InvalidOperationException"></exception>
        public List<Issue> GetTopStaked(DbServiceContext dbServiceContext, PaginatedList paginatedList = null, string userId = null)
        {
            if (_topStakedIssuesPercent == 0)
            {
                throw new InvalidOperationException(Resource.ErrorTopStakePercentNeedsToBeSet);
            }

            List<Issue> issues = GetAll(dbServiceContext, null, userId);

            if (_topStakedIssuesPercent >= 100)
            {
                return issues;
            }

            decimal count = Convert.ToDecimal(issues.Count);

            int limit = Convert.ToInt32(Math.Round(_topStakedIssuesPercent / 100 * count));

            List<Issue> issuesProcessed = issues.Take(limit).ToList();

            issuesProcessed = ProcessPaginatedList(paginatedList, issuesProcessed);

            return issuesProcessed;
        }

        /// <summary>
        /// Gets the issues by tags.
        /// </summary>
        /// <param name="dbServiceContext">The database service context.</param>
        /// <param name="tags">The tags.</param>
        /// <param name="paginatedList">The paginated list.</param>
        /// <param name="userId">The user identifier.</param>
        /// <returns>
        /// All the issues that contain to at least one of the given tags
        /// </returns>
        public List<Issue> GetByTags(DbServiceContext dbServiceContext, string tags, PaginatedList paginatedList = null, string userId = null)
        {
            List<Issue> issues = GetAll(dbServiceContext, null, userId);

            return GetByTags(tags, issues, paginatedList);
        }

        /// <summary>
        /// Gets the top stakes issues percentage by tags.
        /// </summary>
        /// <param name="dbServiceContext">The database service context.</param>
        /// <param name="tags">The tags.</param>
        /// <param name="paginatedList">The paginated list.</param>
        /// <param name="userId">The user identifier.</param>
        /// <returns>
        /// The top staked issues in percent by tags
        /// </returns>
        /// <exception cref="System.InvalidOperationException"></exception>
        public List<Issue> GetTopStakedByTags(DbServiceContext dbServiceContext, string tags, PaginatedList paginatedList = null, string userId = null)
        {
            if (_topStakedIssuesPercent == 0)
            {
                throw new InvalidOperationException(Resource.ErrorTopStakePercentNeedsToBeSet);
            }

            List<Issue> issues = GetTopStaked(dbServiceContext, null, userId);

            return GetByTags(tags, issues, paginatedList);
        }

        /// <summary>
        /// Gets the by identifier.
        /// </summary>
        /// <param name="dbServiceContext">The database service context.</param>
        /// <param name="id">The identifier.</param>
        /// <param name="userId">The user identifier.</param>
        /// <returns>
        /// The issue by id or null if not found
        /// </returns>
        public Issue GetById(DbServiceContext dbServiceContext, string id, string userId = null)
        {
            Issue issue = dbServiceContext.Issues
                .Include(i => i.Proposals)
                .FirstOrDefault(i => i.Id.ToString() == id);

            if (issue != null)
            {
                ProposalService.UpdateStakes(dbServiceContext, issue.Proposals);
                ProposalService.SetHasMyStake(dbServiceContext, issue, userId);
            }

            return issue;
        }

        /// <summary>
        /// Adds the issue.
        /// </summary>
        /// <param name="dbServiceContext">The database service context.</param>
        /// <param name="issueSubmission">The issue submission.</param>
        /// <returns>
        /// The guid of the issue added
        /// </returns>
        /// <exception cref="System.InvalidOperationException">Will be thrown if user is not found</exception>
        public Guid Add(DbServiceContext dbServiceContext, IssueSubmission issueSubmission)
        {
            Issue issue = issueSubmission.ToIssue();

            (bool valid, string errorMessage) = IsValid(issue);

            if (!valid)
            {
                throw new InvalidOperationException(errorMessage);
            }

            Guid userId = issueSubmission.UserId;

            TransactionTypeService transactionTypeService = new TransactionTypeService();
            TransactionType transactionType = transactionTypeService.GetTransactionType(dbServiceContext, TransactionTypeNames.AddIssue);

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
                TransactionId = issue.Id
            };

            wallet.TotalBalance += walletTransaction.Balance;
            dbServiceContext.WalletTransactions.Add(walletTransaction);
            
            issue.Proposals ??= new List<Proposal>();

            dbServiceContext.Issues.Add(issue);
            dbServiceContext.SaveChanges();

            return issue.Id;
        }

        /// <summary>
        /// Updates the specified database service context.
        /// </summary>
        /// <param name="dbServiceContext">The database service context.</param>
        /// <param name="issueSubmission">The issue submission.</param>
        /// <exception cref="System.InvalidOperationException">IssueCannotBeEditedAnymore, ErrorIssueNotFound, ErrorIsNotValid</exception>
        public void Update(DbServiceContext dbServiceContext, IssueSubmission issueSubmission)
        {
            Issue issueToUpdate = dbServiceContext.Issues
                .FirstOrDefault(i => i.Id.ToString() == issueSubmission.Id.ToString());

            if (issueToUpdate == null)
            {
                throw new InvalidOperationException(Resource.ErrorIssueNotFound);
            }

            if (!issueToUpdate.CanEdit(issueSubmission.UserId))
            {
                throw new InvalidOperationException(Resource.IssueCannotBeEditedAnymore);
            }

            issueToUpdate.DueDate = issueSubmission.DueDate;
            issueToUpdate.Tags = issueSubmission.Tags;
            issueToUpdate.Description = issueSubmission.Description;
            issueToUpdate.Title = issueSubmission.Title;

            (bool valid, string errorMessage) = IsValid(issueSubmission.ToIssue());

            if (!valid)
            {
                throw new InvalidOperationException(errorMessage);
            }

            dbServiceContext.Issues.Update(issueToUpdate);
            dbServiceContext.SaveChanges();
        }

        /// <summary>
        /// Gets the tag autocomplete.
        /// </summary>
        /// <param name="dbServiceContext">The database service context.</param>
        /// <param name="tag">The tag.</param>
        /// <returns>The matching tags</returns>
        public List<string> GetTagAutocomplete(DbServiceContext dbServiceContext, string tag)
        {
            if (string.IsNullOrEmpty(tag) ||
                tag.StartsWith("#") && tag.Length < 4 ||
                tag.Length < 3)
            {
                return new List<string>();
            }

            bool hasHashtag = tag.Contains("#");

            string searchTag = tag;

            if (!searchTag.StartsWith("#"))
            {
                searchTag = $"#{tag}";
            }

            List<Issue> issuesWithTag = dbServiceContext.Issues
                .Where(i => i.Tags.Contains(searchTag))
                .ToList();

            List<string> foundTags = new List<string>();

            foreach (Issue issue in issuesWithTag)
            {
                List<string> tags = issue.GetTags().ToList();

                foreach (string tagFromIssue in tags)
                {
                    if (tagFromIssue.ToLowerInvariant().StartsWith(searchTag.ToLowerInvariant()))
                    {
                        foundTags.Add(!hasHashtag ? tagFromIssue.Substring(1): tagFromIssue);
                    }
                }
            }

            foundTags.Sort();

            return foundTags;
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
                        Title = row["Title"].ToString()
                    };

                    string value = row["DueDateDays"].ToString();

                    if (!string.IsNullOrEmpty(value))
                    {
                        issue.DueDate = DateTime.Now.Date.AddDays(Convert.ToInt32(value));
                    }

                    string userId = row["CreatorUserID"].ToString();

                    if (!string.IsNullOrEmpty(userId))
                    {
                        User user = dbServiceContext.Users
                            .FirstOrDefault(u => u.ImportId == Convert.ToInt32(userId));

                        if (user != null)
                        {
                            issue.CreatorUserId = user.Id;
                        }
                    }

                    dbServiceContext.Issues.Add(issue);

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
        /// Gets the issues by tags.
        /// </summary>
        /// <param name="tags">The tags.</param>
        /// <param name="issues">The issues.</param>
        /// <param name="paginatedList">The paginated list.</param>
        /// <returns>The issues from the list containing at least one tag</returns>
        private static List<Issue> GetByTags(string tags, List<Issue> issues, PaginatedList paginatedList)
        {
            List<string> tagsList = new List<string>(Issue.GetTags(tags));

            List<Issue> issuesProcessed = new List<Issue>();

            foreach (Issue issue in issues)
            {
                if (tagsList.Any(tag => issue.HasTag(tag)))
                {
                    issuesProcessed.Add(issue);
                }
            }

            issuesProcessed = ProcessPaginatedList(paginatedList, issuesProcessed);

            return issuesProcessed;
        }

        /// <summary>
        /// Processes the paginated list.
        /// </summary>
        /// <param name="paginatedList">The paginated list.</param>
        /// <param name="issues">The issues.</param>
        /// <returns></returns>
        private static List<Issue> ProcessPaginatedList(PaginatedList paginatedList, List<Issue> issues)
        {
            if (paginatedList is {ItemsPerPage: > 0})
            {
                issues = issues
                    .Skip(paginatedList.Skip)
                    .Take(paginatedList.ItemsPerPage)
                    .ToList();
            }

            if (paginatedList is { GetDetails: false })
            {
                foreach (Issue issue in issues)
                {
                    issue.Proposals = new List<Proposal>();
                }
            }

            return issues;
        }

        /// <summary>
        /// Sets the top staked.
        /// </summary>
        /// <param name="issues">The issues.</param>
        private void SetTopStaked(List<Issue> issues)
        {
            int topStakedIssuesCount = (int)Math.Round(issues.Count * _topStakedIssuesPercent / 100, 0);

            List<Issue> topStakedIssues = issues
                .OrderByDescending(i => i.TotalStakeCount)
                .Take(topStakedIssuesCount)
                .ToList();

            foreach (Issue issue in topStakedIssues)
            {
                if (issue.Proposals.FirstOrDefault(s => s.IsStaked) != null)
                {
                    issue.IsTopStaked = true;
                }
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
