CREATE TABLE [dbo].[Users] (
    [Id]                   UNIQUEIDENTIFIER NOT NULL,
    [ImportId]             INT              NULL,
    [UniqueExternalUserId] UNIQUEIDENTIFIER NOT NULL,
    [UserName]             NVARCHAR (450)   NOT NULL,
    [WalletId]             UNIQUEIDENTIFIER NULL,
    CONSTRAINT [PK_Users] PRIMARY KEY CLUSTERED ([Id] ASC),
    CONSTRAINT [FK_Users_Wallets_WalletId] FOREIGN KEY ([WalletId]) REFERENCES [dbo].[Wallets] ([Id])
);


GO
CREATE NONCLUSTERED INDEX [IX_Users_WalletId]
    ON [dbo].[Users]([WalletId] ASC);


GO
CREATE UNIQUE NONCLUSTERED INDEX [IX_Users_UserName]
    ON [dbo].[Users]([UserName] ASC);


GO
CREATE UNIQUE NONCLUSTERED INDEX [IX_Users_UniqueExternalUserId]
    ON [dbo].[Users]([UniqueExternalUserId] ASC);

