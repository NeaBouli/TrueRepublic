using Common.Entities;

namespace Common.GuiEntities
{
    public class RenderIssue : Issue
    {
        public RenderIssue(Issue issue)
        {
            IsTopStaked = issue.IsTopStaked;
            Tags = issue.Tags;
            CreateDate = issue.CreateDate;
            CreatorUserId = issue.CreatorUserId;
            Description = issue.Description;
            DueDate = issue.DueDate;
            Id = issue.Id;
            ImportId = issue.ImportId;
            Proposals = issue.Proposals;
            Title = issue.Title;

            BackgroundColor = IsTopStaked ? "#e1f5d3" : "#fff2f2";
        }

        public string BackgroundColor { get; set; }
    }
}
