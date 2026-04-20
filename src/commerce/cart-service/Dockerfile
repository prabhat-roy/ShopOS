FROM mcr.microsoft.com/dotnet/sdk:9.0 AS build
WORKDIR /src
COPY *.csproj ./
RUN dotnet restore
COPY . .
RUN dotnet publish CartService.csproj \
    --configuration Release \
    --output /app/publish \
    --no-restore

FROM mcr.microsoft.com/dotnet/aspnet:9.0-noble-chiseled AS final
WORKDIR /app
COPY --from=build /app/publish .
EXPOSE 8130
ENV ASPNETCORE_URLS=http://+:8130
ENV ASPNETCORE_ENVIRONMENT=Production
ENTRYPOINT ["dotnet", "CartService.dll"]
