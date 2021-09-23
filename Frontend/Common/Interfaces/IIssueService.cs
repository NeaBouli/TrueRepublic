using System;
using System.Collections.Generic;
using System.Data;
using Common.Entities;

namespace Common.Interfaces
{
    /// <summary>
    /// Implementation of the IIssues Service interface
    /// </summary>
    public interface IIssueService
    {
        /// <summary>
        /// Gets all issues.
        /// </summary>
        /// <returns>Gets all issues</returns>
        List<Issue> GetAllIssues(bool includeSuggestions = false);

        /// <summary>
        /// Gets all valid issues.
        /// </summary>
        /// <returns>All valid issues that contain at least one stake</returns>
        List<Issue> GetAllValidIssues(bool includeSuggestions = false, bool onlyStaked = false);

        /// <summary>
        /// Gets the top staked issues.
        /// </summary>
        /// <param name="limit">The limit.</param>
        /// <returns>The top staked issues depending on the limit</returns>
        List<Issue> GetTopStakedIssues(int limit);

        /// <summary>
        /// Gets the top staked issues percent.
        /// </summary>
        /// <param name="percentage">The percentage.</param>
        /// <param name="limitNumber">The limit number.</param>
        /// <returns>
        /// The top stacked issues depending on a percentage
        /// </returns>
        List<Issue> GetTopStakedIssuesPercentage(decimal percentage, int limitNumber = 0);

        /// <summary>
        /// Gets the issues by tags.
        /// </summary>
        /// <param name="tags">The tags.</param>
        /// <returns>All the issues that contain to at least one of the given tags</returns>
        List<Issue> GetIssuesByTags(string tags);

        /// <summary>
        /// Gets the top staked issues by tags.
        /// </summary>
        /// <param name="tags">The tags.</param>
        /// <param name="limit">The limit.</param>
        /// <returns>The top staked issues by tags</returns>
        List<Issue> GetTopStakedIssuesByTags(string tags, int limit);

        /// <summary>
        /// Gets the top stakes issues percentage by tags.
        /// </summary>
        /// <param name="tags">The tags.</param>
        /// <param name="percentage">The percentage.</param>
        /// <param name="limitNumber">The limit number.</param>
        /// <returns>
        /// The top staked issues in percent by tags
        /// </returns>
        List<Issue> GetTopStakesIssuesPercentageByTags(string tags, decimal percentage = 100, int limitNumber = 0);

        /// <summary>
        /// Adds the issue.
        /// </summary>
        /// <param name="issue">The issue.</param>
        /// <param name="userId">The user identifier.</param>
        /// <returns>
        /// The guid of the issue added
        /// </returns>
        public Guid AddIssue(Issue issue, Guid userId);

        /// <summary>
        /// Gets the issue.
        /// </summary>
        /// <param name="issueId">The issue unique identifier.</param>
        /// <param name="includeSuggestions">if set to <c>true</c> [with suggestions].</param>
        /// <returns>
        /// The issue for the given issue guid
        /// </returns>
        Issue GetIssue(Guid issueId, bool includeSuggestions = false);

        /// <summary>
        /// Imports the specified data table.
        /// </summary>
        /// <param name="dataTable">The data table.</param>
        /// <returns>The number of imported records</returns>
        public int Import(DataTable dataTable);
    }
}