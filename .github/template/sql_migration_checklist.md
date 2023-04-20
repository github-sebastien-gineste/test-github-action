# Checklist for a PR containing SQL migrations
- [ ] The migrations are [only applied on api-domains MySQL cluster](https://github.com/ebuzzing/service-api-domains/blob/master/documentation/ADR/OnlyApplyMigrationsToApiDomainsCluster.md)
- [ ] I have tested my migration in a sandbox (SQL migrations cannot be tested in a deploy-PR to avoid modifying production's schemas)
- [ ] I will [deploy api-domains-migrations](https://github.com/ebuzzing/service-api-domains/blob/master/documentation/HowTo/HowToDeployApiDomainsMigrations.md) first then `api-domains`