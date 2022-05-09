using System.Collections.Generic;
using System.Linq;
using Microsoft.AspNetCore.Components;
using MudBlazor;

namespace PnyxWebAssembly.Client.Components
{
    public partial class ApiKeyDialog
    {
        /// <summary>
        /// Gets or sets the mud dialog.
        /// </summary>
        /// <value>
        /// The mud dialog.
        /// </value>
        [CascadingParameter]
        private MudDialogInstance MudDialog { get; set; }

        [Parameter]
        public string InfoText { get; set; }

        [Parameter]
        public string Label { get; set; }

        public string Input { get; set; }

        private MudForm Form { get; set; }

        private string[] Errors { get; set; } = { };

        private async void Submit()
        {
            await InvokeAsync(StateHasChanged);

            List<string> errors = ValidateInput(Input).ToList();

            if (errors.Count > 0)
            {
                return;
            }

            MudDialog.Close(DialogResult.Ok(Input));
        }

        /// <summary>
        /// Cancels this instance.
        /// </summary>
        private void Cancel() => MudDialog.Cancel();

        private IEnumerable<string> ValidateInput(string input)
        {
            if (string.IsNullOrWhiteSpace(input))
            {
                yield return "Ein Pixabay API-Key muss angegeben werden";
                yield break;
            }

            if (input.Length < 8)
            {
                yield return "Der Pixabay API-Key muss mindestens 8 Zeichen lang sein";
            }
        }
    }
}
