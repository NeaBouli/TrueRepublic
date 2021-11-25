namespace Common.Entities
{
    /// <summary>
    /// The transaction type names
    /// </summary>
    public enum TransactionTypeNames
    {
        /// <summary>
        /// The genesis transaction
        /// </summary>
        Genesis,

        /// <summary>
        /// The add item
        /// </summary>
        AddIssue,

        /// <summary>
        /// The add Proposal
        /// </summary>
        AddProposal,

        /// <summary>
        /// The stake Proposal
        /// </summary>
        StakeProposal,

        /// <summary>
        /// The stake Proposal rollback
        /// </summary>
        StakeProposalRollback
    }
}
