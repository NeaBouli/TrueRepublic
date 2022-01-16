CREATE TABLE [dbo].[Votes] (
    [Id]               UNIQUEIDENTIFIER NOT NULL,
    [ImportId]         INT              NULL,
    [UserId]           UNIQUEIDENTIFIER NOT NULL,
    [IssueId]          UNIQUEIDENTIFIER NOT NULL,
    [ProposalId]       UNIQUEIDENTIFIER NOT NULL,
    [Value]            INT              NOT NULL,
    [CreateDate]       DATETIME2 (7)    NOT NULL,
    [LastModifiedDate] DATETIME2 (7)    NOT NULL,
    CONSTRAINT [PK_Votes] PRIMARY KEY CLUSTERED ([Id] ASC),
    CONSTRAINT [FK_Votes_Proposals_ProposalId] FOREIGN KEY ([ProposalId]) REFERENCES [dbo].[Proposals] ([Id]) ON DELETE CASCADE,
    CONSTRAINT [FK_Votes_Users_UserId] FOREIGN KEY ([UserId]) REFERENCES [dbo].[Users] ([Id]) ON DELETE CASCADE
);


GO
CREATE NONCLUSTERED INDEX [IX_Votes_UserId]
    ON [dbo].[Votes]([UserId] ASC);


GO
CREATE NONCLUSTERED INDEX [IX_Votes_ProposalId]
    ON [dbo].[Votes]([ProposalId] ASC);

