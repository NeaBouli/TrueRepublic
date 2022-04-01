using Microsoft.AspNetCore.Components;
using MudBlazor;

namespace PnyxWebAssembly.Client.Components
{
    public partial class ImageSelector
    {
        [CascadingParameter] 
        MudDialogInstance MudDialog { get; set; }

        [Parameter] 
        public string Hashtags { get; set; }

        private int SelectedPaper { get; set; }

        private void Submit() => MudDialog.Close(DialogResult.Ok(true));

        private void Cancel() => MudDialog.Cancel();
    }
}
