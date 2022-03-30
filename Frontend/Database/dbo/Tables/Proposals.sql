CREATE TABLE [dbo].[Proposals] (
    [Id]            UNIQUEIDENTIFIER NOT NULL,
    [IssueId]       UNIQUEIDENTIFIER NOT NULL,
    [ImportId]      INT              NULL,
    [Title]         NVARCHAR (MAX)   NOT NULL,
    [Description]   NVARCHAR (MAX)   NOT NULL,
    [CreateDate]    DATETIME2 (7)    NOT NULL,
    [CreatorUserId] UNIQUEIDENTIFIER NOT NULL,
    CONSTRAINT [PK_Proposals] PRIMARY KEY CLUSTERED ([Id] ASC),
    CONSTRAINT [FK_Proposals_Issues_IssueId] FOREIGN KEY ([IssueId]) REFERENCES [dbo].[Issues] ([Id]) ON DELETE CASCADE
);




GO
CREATE NONCLUSTERED INDEX [IX_Proposals_IssueId]
    ON [dbo].[Proposals]([IssueId] ASC);

