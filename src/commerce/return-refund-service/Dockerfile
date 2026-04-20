FROM mcr.microsoft.com/dotnet/sdk:9.0 AS build
WORKDIR /src
COPY *.csproj ./
RUN dotnet restore
COPY . .
RUN dotnet publish ReturnRefundService.csproj \
    --configuration Release \
    --output /app/publish \
    --no-restore

FROM mcr.microsoft.com/dotnet/aspnet:9.0-noble-chiseled AS final
WORKDIR /app
COPY --from=build /app/publish .
EXPOSE 8131
ENV ASPNETCORE_URLS=http://+:8131
ENV ASPNETCORE_ENVIRONMENT=Production
ENTRYPOINT ["dotnet", "ReturnRefundService.dll"]
