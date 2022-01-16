CREATE TABLE [dbo].[TransactionTypes] (
    [Id]       UNIQUEIDENTIFIER NOT NULL,
    [ImportId] INT              NULL,
    [Name]     NVARCHAR (MAX)   NOT NULL,
    [Fee]      FLOAT (53)       NOT NULL,
    CONSTRAINT [PK_TransactionTypes] PRIMARY KEY CLUSTERED ([Id] ASC)
);

