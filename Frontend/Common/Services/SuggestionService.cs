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
        /// <summary>
        /// Gets the suggestions for issue.
        /// </summary>
        /// <param name="dbServiceContext">The database service context.</param>
        /// <param name="issueId">The issue identifier.</param>
        /// <returns></returns>
        private List<Suggestion> GetSuggestionsForIssue(DbServiceContext dbServiceContext, Guid issueId)
        {
            Issue issue = dbServiceContext.Issues
                .Include(i => i.Suggestions)
                .FirstOrDefault(i => i.Id.ToString() == issueId.ToString());

            if (issue == null)
            {
                return null;
            }

            UpdateStakes(dbServiceContext, issue.Suggestions);

            return issue.Suggestions;
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
    }
}
