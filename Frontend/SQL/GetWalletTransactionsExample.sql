/****** Skript für SelectTopNRows-Befehl aus SSMS ******/
SELECT u.[Id]
      ,u.[UserName]
	  ,w.Id WalletId
      ,w.TotalBalance
  FROM [PnyxDb].[dbo].[Users] u
  INNER JOIN Wallets w ON u.WalletId = w.Id
  WHERE u.UserName LIKE 'Arne'

  SELECT w.Id, w.Balance, t.[Name] AS TransactionName, t.Fee, w.CreateDate FROM WalletTransactions w
  INNER JOIN TransactionTypes t ON t.Id = w.TransactionTypeId
  WHERE WalletId = '85A56215-B923-4902-AA0D-2EE138F157FA'
  ORDER BY CreateDate