using System.ComponentModel;
using Microsoft.AspNetCore.Components;
using MudBlazor;
using MudBlazor.Utilities;
using CategoryAttribute = MudBlazor.CategoryAttribute;

namespace PnyxWebAssembly.Client.Components
{
    public partial class RichMudCardMedia : MudComponentBase
    {
        protected string StyleString =>
            StyleBuilder.Default($"background-image:url(\"{Image}\");height: {Height}px;")
                .AddStyle(this.Style)
                .Build();

        protected string Classname =>
            new CssBuilder("mud-card-media")
                .AddClass(Class)
                .Build();

        [Parameter] 
        [Category("Behavior")] 
        public string Title { get; set; }
        
        [Parameter] 
        [Category("Behavior")] 
        public string Image { get; set; }

        [Parameter] 
        [Category("Behavior")] 
        public int Height { get; set; } = 300;

        /// <summary>
        /// Child content of the component.
        /// </summary>
        [Parameter]
        [Category("Behavior")]
        public RenderFragment ChildContent { get; set; }
    }
}
