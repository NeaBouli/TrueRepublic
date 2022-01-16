#See https://aka.ms/containerfastmode to understand how Visual Studio uses this Dockerfile to build your images for faster debugging.

FROM mcr.microsoft.com/dotnet/aspnet:5.0 AS base
WORKDIR /webserviceapp
EXPOSE 81
EXPOSE 444

FROM mcr.microsoft.com/dotnet/sdk:5.0 AS build
WORKDIR /src
COPY ["WebService/WebService.csproj", "WebService/"]
COPY ["Common/Common.csproj", "Common/"]
RUN dotnet restore "WebService/WebService.csproj"
COPY . .
WORKDIR "/src/WebService"
RUN dotnet build "WebService.csproj" -c Release -o /webserviceapp/build

FROM build AS publish
RUN dotnet publish "WebService.csproj" -c Release -o /webserviceapp/publish

FROM base AS final
WORKDIR /webserviceapp
COPY --from=publish /webserviceapp/publish .
ENTRYPOINT ["dotnet", "WebService.dll"]