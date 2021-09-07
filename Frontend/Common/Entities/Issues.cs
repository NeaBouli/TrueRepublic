using System;
using System.Collections.Generic;
using System.Linq;

namespace Common.Entities
{
    public class Issues
    {
        /// <summary>
        /// Gets all issues.
        /// </summary>
        /// <value>
        /// All issues.
        /// </value>
        public List<Issue> AllIssues => new List<Issue>();

        /// <summary>
        /// Gets all valid issues.
        /// </summary>
        /// <returns>All valid issues that contain at least one stake</returns>
        public IEnumerable<Issue> GetAllValidIssues()
        {
            return AllIssues
                .Where(issue => issue.DueDate == null || issue.DueDate >= DateTime.Now)
                .Where(issue => issue.Suggestions.Any(suggestion => suggestion.IsStaked)).ToList();
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

    }
}
