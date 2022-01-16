CREATE TABLE [dbo].[WalletTransactions] (
    [Id]                UNIQUEIDENTIFIER NOT NULL,
    [WalletId]          UNIQUEIDENTIFIER NOT NULL,
    [ImportId]          INT              NULL,
    [Balance]           FLOAT (53)       NOT NULL,
    [TransactionTypeId] UNIQUEIDENTIFIER NOT NULL,
    [TransactionId]     UNIQUEIDENTIFIER NULL,
    [CreateDate]        DATETIME2 (7)    NOT NULL,
    CONSTRAINT [PK_WalletTransactions] PRIMARY KEY CLUSTERED ([Id] ASC),
    CONSTRAINT [FK_WalletTransactions_TransactionTypes_TransactionTypeId] FOREIGN KEY ([TransactionTypeId]) REFERENCES [dbo].[TransactionTypes] ([Id]) ON DELETE CASCADE,
    CONSTRAINT [FK_WalletTransactions_Wallets_WalletId] FOREIGN KEY ([WalletId]) REFERENCES [dbo].[Wallets] ([Id]) ON DELETE CASCADE
);


GO
CREATE NONCLUSTERED INDEX [IX_WalletTransactions_WalletId]
    ON [dbo].[WalletTransactions]([WalletId] ASC);


GO
CREATE NONCLUSTERED INDEX [IX_WalletTransactions_TransactionTypeId]
    ON [dbo].[WalletTransactions]([TransactionTypeId] ASC);

