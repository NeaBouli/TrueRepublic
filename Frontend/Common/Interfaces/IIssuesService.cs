﻿using System;
using System.Collections.Generic;
using System.Data;
using Common.Entities;

namespace Common.Interfaces
{
    /// <summary>
    /// Implementation of the IIssues Service interface
    /// </summary>
    public interface IIssuesService
    {
        /// <summary>
        /// Gets all issues.
        /// </summary>
        /// <returns></returns>
        List<Issue> GetAllIssues();

        /// <summary>
        /// Gets all valid issues.
        /// </summary>
        /// <returns>All valid issues that contain at least one stake</returns>
        IEnumerable<Issue> GetAllValidIssues(bool onlyStaked = false);

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
        /// <returns>
        /// The top stacked issues depending on a percentage
        /// </returns>
        List<Issue> GetTopStakedIssuesPercentage(decimal percentage);

        /// <summary>
        /// Gets the issues by tags.
        /// </summary>
        /// <param name="tags">The tags.</param>
        /// <returns>All the issues that contain to at least one of the given tags</returns>
        IEnumerable<Issue> GetIssuesByTags(string tags);

        /// <summary>
        /// Gets the top staked issues by tags.
        /// </summary>
        /// <param name="tags">The tags.</param>
        /// <param name="limit">The limit.</param>
        /// <returns>The top staked issues by tags</returns>
        IEnumerable<Issue> GetTopStakedIssuesByTags(string tags, int limit);

        /// <summary>
        /// Gets the top stakes issues percentage by tags.
        /// </summary>
        /// <param name="tags">The tags.</param>
        /// <param name="percentage">The percentage.</param>
        /// <returns>The top staked issues in percent by tags</returns>
        List<Issue> GetTopStakesIssuesPercentageByTags(string tags, decimal percentage = 100);

        /// <summary>
        /// Adds the issue.
        /// </summary>
        /// <param name="issue">The issue.</param>
        /// <param name="user">The user.</param>
        /// <returns>
        /// The guid of the issue added
        /// </returns>
        public Guid AddIssue(Issue issue, User user);

        /// <summary>
        /// Gets the issue.
        /// </summary>
        /// <param name="issueGuid">The issue unique identifier.</param>
        /// <returns>The issue for the given issue guid</returns>
        Issue GetIssue(Guid issueGuid);

        public void Import(DataTable dataTable);
    }
}