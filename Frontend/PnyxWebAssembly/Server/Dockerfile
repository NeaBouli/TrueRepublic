#See https://aka.ms/containerfastmode to understand how Visual Studio uses this Dockerfile to build your images for faster debugging.

FROM mcr.microsoft.com/dotnet/aspnet:6.0 AS base
WORKDIR /app
EXPOSE 80
EXPOSE 443

FROM mcr.microsoft.com/dotnet/sdk:6.0 AS build
WORKDIR /src
COPY ["PnyxWebAssembly/Server/PnyxWebAssembly.Server.csproj", "PnyxWebAssembly/Server/"]
COPY ["Common/Common.csproj", "Common/"]
COPY ["PnyxWebAssembly/Client/PnyxWebAssembly.Client.csproj", "PnyxWebAssembly/Client/"]
RUN dotnet restore "PnyxWebAssembly/Server/PnyxWebAssembly.Server.csproj"
COPY . .
WORKDIR "/src/PnyxWebAssembly/Server"
RUN dotnet build "PnyxWebAssembly.Server.csproj" -c Release -o /app/build

FROM build AS publish
RUN dotnet publish "PnyxWebAssembly.Server.csproj" -c Release -o /app/publish

FROM base AS final
WORKDIR /app
COPY --from=publish /app/publish .
ENTRYPOINT ["dotnet", "PnyxWebAssembly.Server.dll"]