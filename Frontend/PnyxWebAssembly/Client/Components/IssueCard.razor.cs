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
        /// Updates the avatar infos.
        /// </summary>
        private async void UpdateAvatarInfos()
        {
            UserCacheService.ClientFactory = ClientFactory;
            User user = await UserCacheService.GetUserById(Issue.CreatorUserId);

            if (user != null)
            {
                CreatorUserName = user.UserName;
            }

            AvatarName = CreatorUserName.Substring(0, 1).ToUpperInvariant();

            AvatarImageCacheService.ClientFactory = ClientFactory;
            HasAvatarImage = await AvatarImageCacheService.HasAvatarImage(CreatorUserName);
            AvatarImage = await AvatarImageCacheService.GetAvatarImageBase64(CreatorUserName);

            await InvokeAsync(StateHasChanged);
        }

        /// <summary>
        /// Updates the image infos.
        /// </summary>
        private async void UpdateImageInfos()
        {
            IssueImageCacheService.ClientFactory = ClientFactory;
            IssueImage = await IssueImageCacheService.GetImageForHashtags(Issue.Tags);

            await InvokeAsync(StateHasChanged);
        }

        /// <summary>
        /// Fixes the properties for card display.
        /// </summary>
        private void FixPropertiesForCardDisplay()
        {
            if (!string.IsNullOrEmpty(_issue.Title) && _issue.Title.Length > 1)
            {
                _issue.Title = $"{_issue.Title.Substring(0, 1).ToUpperInvariant()}{_issue.Title.Substring(1)}";
            }

            if (!string.IsNullOrEmpty(_issue.Description) && _issue.Description.Length > 200)
            {
                string description = _issue.Description.Substring(0, 200);
                int lastIndex = description.LastIndexOf(" ", StringComparison.Ordinal);
                if (lastIndex > 0)
                {
                    description = description.Substring(0, lastIndex);
                    description = $"{description}...";
                }

                _issue.Description = description;
            }
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
