using System;
using System.Net.Http;
using Common.Entities;
using Microsoft.AspNetCore.Components;
using PnyxWebAssembly.Client.Services;

namespace PnyxWebAssembly.Client.Components
{
    /// <summary>
    /// Implementation of the issue card
    /// </summary>
    /// <seealso cref="Microsoft.AspNetCore.Components.ComponentBase" />
    public partial class IssueCard
    {
        /// <summary>
        /// The issue
        /// </summary>
        private Issue _issue = null;

        /// <summary>
        /// Gets or sets the client factory.
        /// </summary>
        /// <value>
        /// The client factory.
        /// </value>
        [Inject]
        private IHttpClientFactory ClientFactory { get; set; }

        /// <summary>
        /// Gets or sets the navigation manager.
        /// </summary>
        /// <value>
        /// The navigation manager.
        /// </value>
        [Inject]
        private NavigationManager NavigationManager { get; set; }

        /// <summary>
        /// Gets or sets the issue.
        /// </summary>
        /// <value>
        /// The issue.
        /// </value>
        [Parameter]
        public Issue Issue
        {
            get => _issue;
            set
            {
                _issue = value;

                if (_issue == null)
                {
                    return;
                }

                FixPropertiesForCardDisplay();

                UpdateAvatarInfos();

                UpdateImageInfos();

                InvokeAsync(StateHasChanged);
            }
        }

        /// <summary>
        /// Gets or sets the name of the user.
        /// </summary>
        /// <value>
        /// The name of the user.
        /// </value>
        [Parameter]
        public string UserName { get; set; } = "";

        /// <summary>
        /// Gets or sets the name of the creator user.
        /// </summary>
        /// <value>
        /// The name of the creator user.
        /// </value>
        public string CreatorUserName { get; set; } = "Unknown";

        /// <summary>
        /// Gets or sets a value indicating whether this instance has avatar image.
        /// </summary>
        /// <value>
        ///   <c>true</c> if this instance has avatar image; otherwise, <c>false</c>.
        /// </value>
        public bool HasAvatarImage { get; set; }

        /// <summary>
        /// Gets or sets the avatar image.
        /// </summary>
        /// <value>
        /// The avatar image.
        /// </value>
        public string AvatarImage { get; set; }

        /// <summary>
        /// Gets or sets the name of the avatar.
        /// </summary>
        /// <value>
        /// The name of the avatar.
        /// </value>
        public string AvatarName { get; set; }

        /// <summary>
        /// Gets or sets the issue image.
        /// </summary>
        /// <value>
        /// The issue image.
        /// </value>
        public string IssueImage { get; set; }

        /// <summary>
        /// Gets or sets the title.
        /// </summary>
        /// <value>
        /// The title.
        /// </value>
        public string Title { get; set; }

        /// <summary>
        /// Gets or sets the tags.
        /// </summary>
        /// <value>
        /// The tags.
        /// </value>
        public string Tags { get; set; }

        /// <summary>
        /// Gets or sets the description.
        /// </summary>
        /// <value>
        /// The description.
        /// </value>
        public string Description { get; set; }

        /// <summary>
        /// Updates the avatar infos.
        /// </summary>
        private async void UpdateAvatarInfos()
        {
            UserService.ClientFactory = ClientFactory;
            User user = await UserService.GetUserById(Issue.CreatorUserId);

            if (user != null)
            {
                CreatorUserName = user.UserName;
            }

            AvatarName = CreatorUserName.Substring(0, 1).ToUpperInvariant();

            AvatarImageService.ClientFactory = ClientFactory;
            AvatarImage = await AvatarImageService.GetAvatarImageBase64(CreatorUserName);

            HasAvatarImage = !string.IsNullOrEmpty(AvatarImage);

            await InvokeAsync(StateHasChanged);
        }

        /// <summary>
        /// Updates the image infos.
        /// </summary>
        private async void UpdateImageInfos()
        {
            IssueImageService.ClientFactory = ClientFactory;
            IssueImage = await IssueImageService.GetImageFromService(Issue.Id);

            await InvokeAsync(StateHasChanged);
        }

        /// <summary>
        /// Fixes the properties for card display.
        /// </summary>
        private void FixPropertiesForCardDisplay()
        {
            int length = 100;

            Title = _issue.Title;

            if (!string.IsNullOrEmpty(Title) && Title.Length > 1)
            {
                Title = $"{Title.Substring(0, 1).ToUpperInvariant()}{Title.Substring(1)}";

                if (Title.Length > length)
                {
                    Title = CutText(Title, length);
                }
            }

            Tags = _issue.Tags;

            if (!string.IsNullOrEmpty(Tags))
            {
                if (Tags.Length > length)
                {
                    Tags = CutText(Tags, length);
                }
            }

            Description = _issue.Description;

            if (!string.IsNullOrEmpty(Description))
            {
                if (Description.Length > length)
                {
                    Description = CutText(Description, length);
                }
            }
        }

        private string CutText(string text, int length)
        {
            text = text.Substring(0, length);
            int lastIndex = text.LastIndexOf(" ", StringComparison.Ordinal);
            if (lastIndex > 0)
            {
                text = text.Substring(0, lastIndex);
                text = $"{text}...";
            }

            return text;
        }

        /// <summary>
        /// Goto the issue details.
        /// </summary>
        private void GotoIssueDetails()
        {
            NavigationManager.NavigateTo($"issue/{Issue.Id}");
        }
    }
}
