CREATE TABLE [dbo].[StakedProposals] (
    [Id]             UNIQUEIDENTIFIER NOT NULL,
    [CreateDate]     DATETIME2 (7)    NOT NULL,
    [ImportId]       INT              NULL,
    [IssueId]        UNIQUEIDENTIFIER NOT NULL,
    [UserId]         UNIQUEIDENTIFIER NOT NULL,
    [ProposalId]     UNIQUEIDENTIFIER NOT NULL,
    [ExpirationDays] INT              NOT NULL,
    CONSTRAINT [PK_StakedProposals] PRIMARY KEY CLUSTERED ([Id] ASC),
    CONSTRAINT [FK_StakedProposals_Proposals_ProposalId] FOREIGN KEY ([ProposalId]) REFERENCES [dbo].[Proposals] ([Id]) ON DELETE CASCADE,
    CONSTRAINT [FK_StakedProposals_Users_UserId] FOREIGN KEY ([UserId]) REFERENCES [dbo].[Users] ([Id]) ON DELETE CASCADE
);


GO
CREATE NONCLUSTERED INDEX [IX_StakedProposals_UserId]
    ON [dbo].[StakedProposals]([UserId] ASC);


GO
CREATE NONCLUSTERED INDEX [IX_StakedProposals_ProposalId]
    ON [dbo].[StakedProposals]([ProposalId] ASC);

