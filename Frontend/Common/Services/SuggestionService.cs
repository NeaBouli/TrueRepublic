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
        /// <param name="issueId">The issue identifier.</param>
        /// <returns></returns>
        public List<Suggestion> GetSuggestionsForIssue(Guid issueId)
        {
            DbServiceContext dbServiceContext = DatabaseInitializationService.GetDbServiceContext();

            using (dbServiceContext)
            {
                Issue issue = dbServiceContext.Issues
                    .Include(i => i.Suggestions)
                    .FirstOrDefault(i => i.Id.ToString() == issueId.ToString());

                if (issue == null)
                {
                    return null;
                }

                UpdateStakes(issue.Suggestions);

                return issue.Suggestions;
            }
        }

        /// <summary>
        /// Updates the stakes.
        /// </summary>
        /// <param name="suggestions">The suggestions.</param>
        public static void UpdateStakes(List<Suggestion> suggestions)
        {
            DbServiceContext dbServiceContext = DatabaseInitializationService.GetDbServiceContext();

            using (dbServiceContext)
            {
                foreach (Suggestion suggestion in suggestions)
                {
                    int count = dbServiceContext.StakedSuggestions
                        .Include(s => s.Suggestion)
                        .Where(s => s.Suggestion.Id.ToString() == suggestion.Id.ToString())
                        .ToList().Count;

                    suggestion.StakeCount = count;
                }
            }
        }

        public int Import(DataTable dataTable)
        {
            throw new NotImplementedException();
        }
    }
}
