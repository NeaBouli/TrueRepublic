CREATE TABLE [dbo].[Issues] (
    [Id]            UNIQUEIDENTIFIER NOT NULL,
    [ImportId]      INT              NULL,
    [Tags]          NVARCHAR (MAX)   NOT NULL,
    [Title]         NVARCHAR (MAX)   NOT NULL,
    [Description]   NVARCHAR (MAX)   NOT NULL,
    [DueDate]       DATETIME2 (7)    NULL,
    [CreateDate]    DATETIME2 (7)    NOT NULL,
    [CreatorUserId] UNIQUEIDENTIFIER NOT NULL,
    CONSTRAINT [PK_Issues] PRIMARY KEY CLUSTERED ([Id] ASC)
);

