using System;
using System.IO;
using System.Runtime.InteropServices;
using System.Threading;
using Common.Data;
using Common.Entities;
using Common.Services;
using Microsoft.AspNetCore.Authentication;
using Microsoft.AspNetCore.Builder;
using Microsoft.AspNetCore.Hosting;
using Microsoft.EntityFrameworkCore;
using Microsoft.Extensions.Configuration;
using Microsoft.Extensions.DependencyInjection;
using Microsoft.Extensions.Hosting;
using Microsoft.Extensions.Logging;
using PnyxWebAssembly.Server.Data;
using PnyxWebAssembly.Server.Models;

namespace PnyxWebAssembly.Server
{
    public class Startup
    {
        public Startup(IConfiguration configuration)
        {
            Configuration = configuration;
        }

        public IConfiguration Configuration { get; }

        // This method gets called by the runtime. Use this method to add services to the container.
        // For more information on how to configure your application, visit https://go.microsoft.com/fwlink/?LinkID=398940
        public void ConfigureServices(IServiceCollection services)
        {
            string authConnectString = Configuration.GetConnectionString("DefaultConnection");

            string authDockerConnectString = Environment.GetEnvironmentVariable("DBCONNECTSTRING_AUTH");

            if (!string.IsNullOrEmpty(authDockerConnectString))
            {
                authConnectString = authDockerConnectString;
            }

            DatabaseInitializationService.Platform = Platform.Windows;

            DatabaseInitializationService.DbAuthConnectString = authConnectString;

            string pnyxConnectString = Configuration["DBConnectString"];

            string pnyxDockerConnectString = Environment.GetEnvironmentVariable("DBCONNECTSTRING_PNYX");

            if (!string.IsNullOrEmpty(pnyxDockerConnectString))
            {
                pnyxConnectString = pnyxDockerConnectString;
                DatabaseInitializationService.Platform = Platform.Docker;
            }
            else
            {
                if (RuntimeInformation.IsOSPlatform(OSPlatform.OSX))
                {
                    DatabaseInitializationService.Platform = Platform.Mac;
                    authConnectString = Configuration.GetConnectionString("DefaultConnectionMac");
                    pnyxConnectString = Configuration["DBConnectStringMac"];
                }
            }


            DatabaseInitializationService.DbConnectString = pnyxConnectString;
            DbServiceContext.ConnectString = pnyxConnectString;

            services.AddDbContext<ApplicationDbContext>(options =>
                options.UseSqlServer(authConnectString));

            services.AddDbContext<DbServiceContext>(options =>
                options.UseSqlServer(pnyxConnectString));

            services.AddDatabaseDeveloperPageExceptionFilter();

            services.AddDefaultIdentity<ApplicationUser>(options => options.SignIn.RequireConfirmedAccount = true)
                .AddEntityFrameworkStores<ApplicationDbContext>();

            services.AddIdentityServer()
                .AddApiAuthorization<ApplicationUser, ApplicationDbContext>();

            services.AddAuthentication()
                .AddIdentityServerJwt();

            services.AddControllersWithViews();
            services.AddRazorPages();
        }

        // This method gets called by the runtime. Use this method to configure the HTTP request pipeline.
        public void Configure(
            IApplicationBuilder app, 
            IWebHostEnvironment env, 
            ApplicationDbContext applicationDbContext, 
            DbServiceContext dbServiceContext,
            ILogger<Startup> logger)
        {
            if (env.IsDevelopment())
            {
                app.UseDeveloperExceptionPage();
                app.UseMigrationsEndPoint();
                app.UseWebAssemblyDebugging();
            }
            else
            {
                app.UseExceptionHandler("/Error");
                // The default HSTS value is 30 days. You may want to change this for production scenarios, see https://aka.ms/aspnetcore-hsts.
                app.UseHsts();
            }

            InitDatabase(applicationDbContext, dbServiceContext, logger);

            ImportExcelFile(logger);

            app.UseHttpsRedirection();
            app.UseBlazorFrameworkFiles();
            app.UseStaticFiles();

            app.UseRouting();

            app.UseIdentityServer();
            app.UseAuthentication();
            app.UseAuthorization();

            app.UseEndpoints(endpoints =>
            {
                endpoints.MapRazorPages();
                endpoints.MapControllers();
                endpoints.MapFallbackToFile("index.html");
            });
        }

        /// <summary>
        /// Initializes the database.
        /// </summary>
        /// <param name="applicationDbContext">The application database context.</param>
        /// <param name="dbServiceContext">The database service context.</param>
        /// <param name="logger">The logger.</param>
        private void InitDatabase(ApplicationDbContext applicationDbContext, DbServiceContext dbServiceContext, ILogger<Startup> logger)
        {
            logger.LogInformation($"Init database. Platform is {DatabaseInitializationService.Platform}");

            string sourceInfo = DatabaseInitializationService.Platform == Platform.Docker
                ? "container environment variable DBCONNECTSTRING_AUTH"
                : "appsettings.json";

            try
            {
                logger.LogInformation(
                    $"Trying to connect to PnyxAuthenticationDB as configured in {sourceInfo}: " +
                    $"{DatabaseInitializationService.DbAuthConnectString}");

                if (DatabaseInitializationService.Platform == Platform.Docker)
                {
                    logger.LogInformation("Waiting 10 Seconds for docker DB to come up");
                    Thread.Sleep(10000);
                }

                applicationDbContext.Database.Migrate();

                sourceInfo = DatabaseInitializationService.Platform == Platform.Docker
                    ? "container environment variable DBCONNECTSTRING_PNYX"
                    : "appsettings.json";

                logger.LogInformation(
                    $"Trying to connect to PnyxAuthenticationDB as configured in {sourceInfo}: " +
                    $"{DatabaseInitializationService.DbConnectString}");

                dbServiceContext.Database.EnsureCreated();
            }
            catch (Exception)
            {
                sourceInfo = DatabaseInitializationService.Platform == Platform.Docker
                    ? "container environment variables"
                    : "appsettings.json";

                logger.LogError($"Error could not find database that was defined in {sourceInfo}");

                if (DatabaseInitializationService.Platform == Platform.Docker)
                {
                    logger.LogError("Make sure that 'docker-compose up is used from compose file found here: " +
                                    "https://github.com/NeaBouli/pnyx/blob/development/Frontend/docker-compose.yml");
                }

                throw;
            }
        }

        /// <summary>
        /// Imports the excel file.
        /// </summary>
        private static void ImportExcelFile(ILogger<Startup> logger)
        {
            logger.LogInformation("Trying to import excel file with initial data TestData.xls...");

            ExcelImporterService excelImporterService = new ExcelImporterService();
            if (excelImporterService.IsDbEmpty())
            {
                string excelImportFile = "TestData.xlsx";

                if (DatabaseInitializationService.Platform == Platform.Docker)
                {
                    if (File.Exists("/app/bin/Debug/net5.0/TestData.xlsx"))
                    {
                        excelImportFile = "/app/bin/Debug/net5.0/TestData.xlsx";
                    }
                }

                logger.LogInformation(!excelImportFile.Contains("/")
                    ? "TestData.xls expected in root application directory"
                    : $"TestData.xls expected in {excelImportFile}");

                excelImporterService.ImportExcelFile(excelImportFile);
            }
        }
    }
}
