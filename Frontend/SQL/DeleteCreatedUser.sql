/****** Script for SelectTopNRows command from SSMS  ******/

DELETE
  FROM [PnyxDb].[dbo].Users
  WHERE ImportId IS NULL

  GO

DELETE
  FROM [PnyxDb].[dbo].[WalletTransactions]
  WHERE ImportId IS NULL

  GO

DELETE
  FROM [PnyxDb].[dbo].[Wallets]
  WHERE ImportId IS NULL

  GO





