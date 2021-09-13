using System;
using System.Collections.Generic;
using System.Data;
using Common.Data;
using Common.Entities;
using Common.Interfaces;

namespace Common.Services
{
    public class IssuesDatabaseService : IIssuesService
    {
        public static string DbConnectString { get; set; }

        public IssuesDatabaseService()
        {
            if (string.IsNullOrEmpty(DbConnectString))
            {
                throw new InvalidOperationException(Resource.ErrorDbConnectStringCannotBeEmpty);
            }
        }

        public List<Issue> GetAllIssues()
        {
            throw new NotImplementedException();
        }

        public IEnumerable<Issue> GetAllValidIssues(bool onlyStaked = false)
        {
            throw new NotImplementedException();
        }

        public List<Issue> GetTopStakedIssues(int limit)
        {
            throw new NotImplementedException();
        }

        public List<Issue> GetTopStakedIssuesPercentage(decimal percentage)
        {
            throw new NotImplementedException();
        }

        public IEnumerable<Issue> GetIssuesByTags(string tags)
        {
            throw new NotImplementedException();
        }

        public IEnumerable<Issue> GetTopStakedIssuesByTags(string tags, int limit)
        {
            throw new NotImplementedException();
        }

        public List<Issue> GetTopStakesIssuesPercentageByTags(string tags, decimal percentage = 100)
        {
            throw new NotImplementedException();
        }

        public Guid AddIssue(Issue issue, User user)
        {
            throw new NotImplementedException();
        }

        public Issue GetIssue(Guid issueGuid)
        {
            throw new NotImplementedException();
        }

        public void Import(DataTable issues)
        {
            throw new NotImplementedException();
        }
    }
}
