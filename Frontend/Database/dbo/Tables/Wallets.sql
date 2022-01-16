CREATE TABLE [dbo].[Wallets] (
    [Id]           UNIQUEIDENTIFIER NOT NULL,
    [ImportId]     INT              NULL,
    [TotalBalance] FLOAT (53)       NOT NULL,
    CONSTRAINT [PK_Wallets] PRIMARY KEY CLUSTERED ([Id] ASC)
);

