CREATE TABLE [dbo].[ImageInfos] (
    [Id]       UNIQUEIDENTIFIER NOT NULL,
    [Hashtags] NVARCHAR (MAX)   NOT NULL,
    [Filename] NVARCHAR (MAX)   NOT NULL,
    CONSTRAINT [PK_ImageInfos] PRIMARY KEY CLUSTERED ([Id] ASC)
);

