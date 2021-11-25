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
    /// Implementation of the Proposal service
    /// </summary>
    public class ProposalService
    {
        private readonly decimal _topStakedProposalsPercent;

        /// <summary>
        /// Initializes a new instance of the <see cref="ProposalService"/> class.
        /// </summary>
        public ProposalService()
        {
            _topStakedProposalsPercent = 0;
        }

        /// <summary>
        /// Initializes a new instance of the <see cref="ProposalService" /> class.
        /// </summary>
        /// <param name="topStakedProposalsPercent">The top staked proposals percent.</param>
        public ProposalService(decimal topStakedProposalsPercent)
        {
            _topStakedProposalsPercent = topStakedProposalsPercent;
        }

        /// <summary>
        /// Gets the by identifier.
        /// </summary>
        /// <param name="dbServiceContext">The database service context.</param>
        /// <param name="id">The identifier.</param>
        /// <param name="userId">The user identifier.</param>
        /// <returns>
        /// Gets the proposals for the given id
        /// </returns>
        /// <exception cref="System.InvalidOperationException"></exception>
        public List<Proposal> GetByIssueId(DbServiceContext dbServiceContext, string id, string userId)
        {
            if (_topStakedProposalsPercent == 0)
            {
                throw new InvalidOperationException(Resource.ErrorTopStakePercentNeedsToBeSet);
            }

            Issue issue = dbServiceContext.Issues
                .Include(i => i.Proposals)
                .FirstOrDefault(i => i.Id.ToString() == id);

            if (issue == null)
            {
                return null;
            }

            issue.Proposals ??= new List<Proposal>();

            UpdateStakes(dbServiceContext, issue.Proposals);

            SetTopStaked(issue.Proposals);

            SetHasMyStake(dbServiceContext, issue, userId);

            UpdateVotes(dbServiceContext, issue.Proposals);

            SetTopVoted(issue.Proposals);

            SetMyVote(dbServiceContext, issue, userId);

            return issue.Proposals.OrderByDescending(s => s.IsTopStaked).ThenBy(s => s.CreateDate).ToList();
        }

        /// <summary>
        /// Gets the by proposal identifier.
        /// </summary>
        /// <param name="dbServiceContext">The database service context.</param>
        /// <param name="id">The identifier.</param>
        /// <param name="userId">The user identifier.</param>
        /// <returns></returns>
        /// <exception cref="System.InvalidOperationException"></exception>
        public Proposal GetByProposalId(DbServiceContext dbServiceContext, string id, string userId = null)
        {
            if (_topStakedProposalsPercent == 0)
            {
                throw new InvalidOperationException(Resource.ErrorTopStakePercentNeedsToBeSet);
            }

            Proposal proposal = dbServiceContext.Proposals
                .FirstOrDefault(s => s.Id.ToString() == id);

            if (proposal == null)
            {
                return null;
            }

            UpdateStakes(dbServiceContext, new List<Proposal> { proposal });

            SetTopStaked(new List<Proposal> { proposal });

            if (!string.IsNullOrEmpty(userId))
            {
                SetHasMyStake(dbServiceContext, proposal, userId);
            }

            UpdateVotes(dbServiceContext, new List<Proposal> { proposal });

            SetTopVoted(new List<Proposal> { proposal });

            if (!string.IsNullOrEmpty(userId))
            {
                SetMyVote(dbServiceContext, proposal, userId);
            }

            return proposal;
        }

        /// <summary>
        /// Sets the has my stake.
        /// </summary>
        /// <param name="dbServiceContext">The database service context.</param>
        /// <param name="issue">The issue.</param>
        /// <param name="userId">The user identifier.</param>
        public static void SetHasMyStake(DbServiceContext dbServiceContext, Issue issue, string userId)
        {
            if (string.IsNullOrEmpty(userId))
            {
                return;
            }

            StakedProposal stakedProposal = dbServiceContext.StakedProposals
                .FirstOrDefault(s => s.IssueId.ToString() == issue.Id.ToString() && s.UserId.ToString() == userId);

            if (stakedProposal != null)
            {
                foreach (var proposal in issue.Proposals
                    .Where(proposal => proposal.Id.ToString() == stakedProposal.ProposalId.ToString()))
                {
                    proposal.HasMyStake = true;
                    break;
                }
            }
        }

        /// <summary>
        /// Sets the has my stake.
        /// </summary>
        /// <param name="dbServiceContext">The database service context.</param>
        /// <param name="proposal">The proposal.</param>
        /// <param name="userId">The user identifier.</param>
        public static void SetHasMyStake(DbServiceContext dbServiceContext, Proposal proposal, string userId)
        {
            if (string.IsNullOrEmpty(userId))
            {
                return;
            }

            StakedProposal stakedProposal = dbServiceContext.StakedProposals
                .FirstOrDefault(s => s.ProposalId.ToString() == proposal.Id.ToString() && s.UserId.ToString() == userId);

            if (stakedProposal != null)
            {
                proposal.HasMyStake = true;
            }
        }

        /// <summary>
        /// Updates the stakes.
        /// </summary>
        /// <param name="dbServiceContext">The database service context.</param>
        /// <param name="proposals">The proposals.</param>
        public static void UpdateStakes(DbServiceContext dbServiceContext, List<Proposal> proposals)
        {
            foreach (Proposal proposal in proposals)
            {
                int count = dbServiceContext.StakedProposals
                    .Where(s => s.ProposalId.ToString() == proposal.Id.ToString())
                    .ToList().Count;

                proposal.StakeCount = count;
            }
        }

        /// <summary>
        /// Sets my vote.
        /// </summary>
        /// <param name="dbServiceContext">The database service context.</param>
        /// <param name="issue">The issue.</param>
        /// <param name="userId">The user identifier.</param>
        public static void SetMyVote(DbServiceContext dbServiceContext, Issue issue, string userId)
        {
            if (string.IsNullOrEmpty(userId))
            {
                return;
            }

            List<Vote> votes = dbServiceContext.Votes
                .Include(v => v.Proposal)
                .Where(v => v.UserId.ToString() == userId && v.IssueId.ToString() == issue.Id.ToString())
                .ToList();

            if (votes.Count == 0)
            {
                return;
            }

            foreach (Proposal proposal in issue.Proposals)
            {
                Vote vote = votes.FirstOrDefault(v => v.ProposalId.ToString() == proposal.Id.ToString());

                if (vote != null)
                {
                    proposal.MyVote = vote.Value;
                }
            }
        }

        /// <summary>
        /// Sets my vote.
        /// </summary>
        /// <param name="dbServiceContext">The database service context.</param>
        /// <param name="proposal">The proposal.</param>
        /// <param name="userId">The user identifier.</param>
        public static void SetMyVote(DbServiceContext dbServiceContext, Proposal proposal, string userId)
        {
            if (string.IsNullOrEmpty(userId))
            {
                return;
            }

            var vote = dbServiceContext.Votes
                .Include(v => v.Proposal)
                .FirstOrDefault(v => v.UserId.ToString() == userId && v.ProposalId.ToString() == proposal.Id.ToString());

            if (vote == null)
            {
                return;
            }

            proposal.MyVote = vote.Value;
        }

        /// <summary>
        /// Updates the votes.
        /// </summary>
        /// <param name="dbServiceContext">The database service context.</param>
        /// <param name="proposals">The proposals.</param>
        public static void UpdateVotes(DbServiceContext dbServiceContext, List<Proposal> proposals)
        {
            foreach (Proposal proposal in proposals)
            {
                int count = dbServiceContext.Votes
                    .Where(v => v.ProposalId.ToString() == proposal.Id.ToString())
                    .ToList().Count;

                proposal.VoteCount = count;
            }
        }

        /// <summary>
        /// Adds the specified database service context.
        /// </summary>
        /// <param name="dbServiceContext">The database service context.</param>
        /// <param name="proposalSubmission">The proposal submission.</param>
        /// <returns></returns>
        /// <exception cref="System.InvalidOperationException">
        /// </exception>
        public Guid Add(DbServiceContext dbServiceContext, ProposalSubmission proposalSubmission)
        {
            Proposal proposal = proposalSubmission.ToProposal();

            (bool valid, string errorMessage) = IsValid(proposal);

            if (!valid)
            {
                throw new InvalidOperationException(errorMessage);
            }

            string issueId = proposalSubmission.IssueId.ToString();

            bool proposalWithSameTitleAlreadyExists = dbServiceContext.Proposals
                .FirstOrDefault(s => s.IssueId.ToString() == issueId &&
                                     string.Equals(s.Title, proposalSubmission.Title, StringComparison.OrdinalIgnoreCase)) != null;

            if (proposalWithSameTitleAlreadyExists)
            {
                throw new InvalidOperationException(Resource.ErrorProposalWithSameTitleAlreadyExists);
            }

            Guid userId = proposalSubmission.UserId;

            TransactionTypeService transactionTypeService = new TransactionTypeService();
            TransactionType transactionType = transactionTypeService.GetTransactionType(dbServiceContext, TransactionTypeNames.AddProposal);

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
                TransactionId = proposal.Id
            };

            wallet.TotalBalance += walletTransaction.Balance;
            dbServiceContext.WalletTransactions.Add(walletTransaction);


            dbServiceContext.Proposals.Add(proposal);
            dbServiceContext.SaveChanges();

            return proposal.Id;
        }

        /// <summary>
        /// Updates the specified database service context.
        /// </summary>
        /// <param name="dbServiceContext">The database service context.</param>
        /// <param name="proposalSubmission">The proposal submission.</param>
        /// <exception cref="System.InvalidOperationException">
        /// </exception>
        public void Update(DbServiceContext dbServiceContext, ProposalSubmission proposalSubmission)
        {
            Proposal proposalToUpdate = dbServiceContext.Proposals
                .FirstOrDefault(s => s.Id.ToString() == proposalSubmission.Id.ToString());

            if (proposalToUpdate == null)
            {
                throw new InvalidOperationException(Resource.ErrorIssueNotFound);
            }

            if (!proposalToUpdate.CanEdit(proposalSubmission.UserId))
            {
                throw new InvalidOperationException(Resource.IssueCannotBeEditedAnymore);
            }

            proposalToUpdate.Description = proposalSubmission.Description;
            proposalToUpdate.Title = proposalSubmission.Title;

            (bool valid, string errorMessage) = IsValid(proposalSubmission.ToProposal());

            if (!valid)
            {
                throw new InvalidOperationException(errorMessage);
            }

            dbServiceContext.Proposals.Update(proposalToUpdate);
            dbServiceContext.SaveChanges();
        }

        /// <summary>
        /// Imports the specified data table.
        /// </summary>
        /// <param name="dataTable">The data table.</param>
        /// <returns>The number of imported records</returns>
        public int Import(DataTable dataTable)
        {
            DbServiceContext dbServiceContext = DatabaseInitializationService.GetDbServiceContext();

            using (dbServiceContext)
            {
                int count = dbServiceContext.Proposals.Count();

                if (count > 0)
                {
                    return 0;
                }

                int recordCount = 0;

                foreach (DataRow row in dataTable.Rows)
                {
                    Proposal proposal = new Proposal
                    {
                        ImportId = Convert.ToInt32(row["ID"].ToString()),
                        Description = row["Description"].ToString(),
                        Title = row["Title"].ToString(),
                    };

                    int importIssueId = Convert.ToInt32(row["IssueId"].ToString());

                    Guid? issueId = dbServiceContext.Issues.FirstOrDefault(i => i.ImportId == importIssueId)?.Id;

                    if (issueId != null)
                    {
                        proposal.IssueId = (Guid)issueId;
                    }

                    string userId = row["CreatorUserID"].ToString();

                    if (!string.IsNullOrEmpty(userId))
                    {
                        User user = dbServiceContext.Users
                            .FirstOrDefault(u => u.ImportId == Convert.ToInt32(userId));

                        if (user != null)
                        {
                            proposal.CreatorUserId = user.Id;
                        }
                    }

                    dbServiceContext.Proposals.Add(proposal);

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
        /// Sets the top staked.
        /// </summary>
        /// <param name="proposals">The proposals.</param>
        private void SetTopStaked(List<Proposal> proposals)
        {
            int topStakedIssuesCount = (int)Math.Round(proposals.Count * _topStakedProposalsPercent / 100, 0);

            List<Proposal> topStakedProposals = proposals
                .OrderByDescending(i => i.StakeCount)
                .Take(topStakedIssuesCount)
                .ToList();

            foreach (var proposal in topStakedProposals)
            {
                proposal.IsTopStaked = true;
            }
        }

        /// <summary>
        /// Sets the top voted.
        /// </summary>
        /// <param name="proposals">The proposals.</param>
        private void SetTopVoted(List<Proposal> proposals)
        {
            int topVotedCount = (int)Math.Round(proposals.Count * _topStakedProposalsPercent / 100, 0);

            List<Proposal> topVotedProposals = proposals
                .OrderByDescending(i => i.VoteCount)
                .Take(topVotedCount)
                .ToList();

            foreach (var proposal in topVotedProposals)
            {
                proposal.IsTopVoted = true;
            }
        }

        /// <summary>
        /// Returns true if ... is valid.
        /// </summary>
        /// <param name="proposal">The proposal.</param>
        /// <returns>
        ///   <c>true</c> if this instance is valid; otherwise, <c>false</c>.
        /// </returns>
        private (bool, string) IsValid(Proposal proposal)
        {
            string errorMessage = string.Empty;

            if (string.IsNullOrEmpty(proposal.Title))
            {
                errorMessage = Resource.ErrorTitleIsRequired;
                return (false, errorMessage);
            }

            if (string.IsNullOrEmpty(proposal.Description))
            {
                errorMessage = Resource.ErrorDescriptionIsRequired;
                return (false, errorMessage);
            }

            if (!string.IsNullOrEmpty(proposal.Title) && proposal.Title.Length < 5)
            {
                errorMessage = Resource.ErrorTitleNotLongEnough;
                return (false, errorMessage);
            }

            if (!string.IsNullOrEmpty(proposal.Description) && proposal.Description.Length < 5)
            {
                errorMessage = Resource.ErrorDescriptionNotLongEnough;
                return (false, errorMessage);
            }

            return (true, errorMessage);
        }
    }
}
